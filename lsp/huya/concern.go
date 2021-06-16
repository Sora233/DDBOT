package huya

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	"reflect"
	"runtime"
)

var logger = utils.GetModuleLogger("huya-concern")

type EventType int64

const (
	Live EventType = iota
)

type ConcernEvent interface {
	Type() EventType
}

type ConcernLiveNotify struct {
	LiveInfo
	GroupCode int64 `json:"group_code"`
}

func (notify *ConcernLiveNotify) ShouldSend() bool {
	return notify.Living
}
func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}
func (notify *ConcernLiveNotify) GetUid() interface{} {
	return notify.RoomId
}

func (notify *ConcernLiveNotify) ToMessage() []message.IMessageElement {
	var result []message.IMessageElement
	if notify.Living {
		result = append(result, localutils.MessageTextf("虎牙-%s正在直播【%v】\n%v", notify.Name, notify.RoomName, notify.RoomUrl))
		cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.Avatar, false, proxy_pool.PreferNone)
		if err != nil {
			logger.WithField("avatar", notify.Avatar).Errorf("upload avatar failed %v", err)
		} else {
			result = append(result, cover)
		}
	} else {
		result = append(result, localutils.MessageTextf("虎牙-%s直播结束了\n%v", notify.Name, notify.RoomUrl))
	}
	return result
}

func (notify *ConcernLiveNotify) Type() concern.Type {
	return concern.HuyaLive
}

func NewConcernLiveNotify(groupCode int64, l *LiveInfo) *ConcernLiveNotify {
	if l == nil {
		return nil
	}
	return &ConcernLiveNotify{
		*l,
		groupCode,
	}
}

type Concern struct {
	*StateManager

	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
	stop      chan interface{}
	stopped   bool
}

func (c *Concern) Start() {

	err := c.StateManager.Start()
	if err != nil {
		logger.Errorf("state manager start err %v", err)
	}

	if runtime.NumCPU() >= 3 {
		for i := 0; i < 3; i++ {
			go c.notifyLoop()
		}
	} else {
		go c.notifyLoop()
	}

	go c.EmitFreshCore("huya", func(ctype concern.Type, id interface{}) error {
		roomid, ok := id.(string)
		if !ok {
			return fmt.Errorf("cast fresh id type<%v> to string failed", reflect.ValueOf(id).Type().String())
		}
		if ctype.ContainAll(concern.HuyaLive) {
			oldInfo, _ := c.FindRoom(roomid, false)
			liveInfo, err := c.FindRoom(roomid, true)
			if err != nil {
				return fmt.Errorf("load liveinfo failed %v", err)
			}
			if oldInfo != nil && oldInfo.Living != liveInfo.Living {
				liveInfo.LiveStatusChanged = true
			}
			if oldInfo == nil || oldInfo.Living != liveInfo.Living || oldInfo.RoomName != liveInfo.RoomName {
				c.eventChan <- liveInfo
			}
		}
		return nil
	})
}

func (c *Concern) Add(groupCode int64, id interface{}, ctype concern.Type) (*LiveInfo, error) {
	var err error
	log := logger.WithField("GroupCode", groupCode).WithField("id", id)

	err = c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}

	liveInfo, err := RoomPage(id.(string))
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("查询房间信息失败 %v - %v", id, err)
	}
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return liveInfo, nil
}

func (c *Concern) ListWatching(groupCode int64, ctype concern.Type) ([]*LiveInfo, []concern.Type, error) {
	log := logger.WithField("group_code", groupCode)

	ids, ctypes, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(ctype)
	})
	if err != nil {
		return nil, nil, err
	}
	var resultTypes = make([]concern.Type, 0, len(ids))
	var result = make([]*LiveInfo, 0, len(ids))
	for index, id := range ids {
		liveInfo, err := c.FindOrLoadRoom(id.(string))
		if err != nil {
			log.WithField("id", id).Errorf("get LiveInfo err %v", err)
			continue
		}
		result = append(result, liveInfo)
		resultTypes = append(resultTypes, ctypes[index])
	}

	return result, resultTypes, nil
}

func (c *Concern) notifyLoop() {
	for ievent := range c.eventChan {
		if c.stopped {
			return
		}
		switch ievent.Type() {
		case Live:
			event := ievent.(*LiveInfo)
			log := logger.WithField("name", event.Name).
				WithField("roomid", event.RoomId).
				WithField("title", event.RoomName).
				WithField("living", event.Living)
			log.Debugf("debug event")

			groups, _, _, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern.Type) bool {
				return id.(string) == event.RoomId && p.ContainAny(concern.HuyaLive)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}
			for _, groupCode := range groups {
				notify := NewConcernLiveNotify(groupCode, event)
				c.notify <- notify
				if event.Living {
					log.Debug("living notify")
				} else {
					log.Debug("noliving notify")
				}
			}
		}
	}
}

func (c *Concern) FindRoom(roomId string, load bool) (*LiveInfo, error) {
	var liveInfo *LiveInfo
	if load {
		var err error
		liveInfo, err = RoomPage(roomId)
		if err != nil {
			return nil, err
		}
		_ = c.StateManager.AddLiveInfo(liveInfo)
	}
	if liveInfo != nil {
		return liveInfo, nil
	}
	return c.StateManager.GetLiveInfo(roomId)
}

func (c *Concern) FindOrLoadRoom(roomId string) (*LiveInfo, error) {
	info, _ := c.FindRoom(roomId, false)
	if info == nil {
		return c.FindRoom(roomId, true)
	}
	return info, nil
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		StateManager: NewStateManager(),
		eventChan:    make(chan ConcernEvent, 500),
		notify:       notify,
		stop:         make(chan interface{}),
	}
	return c
}
