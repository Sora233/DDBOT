package douyu

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"reflect"
	"runtime"
	"sync"
)

var logger = utils.GetModuleLogger("douyu-concern")

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

func (notify *ConcernLiveNotify) Type() concern.Type {
	return concern.DouyuLive
}
func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}
func (notify *ConcernLiveNotify) GetUid() interface{} {
	return notify.RoomId
}

func (notify *ConcernLiveNotify) ToMessage() []message.IMessageElement {
	var result []message.IMessageElement
	switch notify.ShowStatus {
	case ShowStatus_Living:
		result = append(result, localutils.MessageTextf("斗鱼-%s正在直播【%v】\n%v", notify.Nickname, notify.RoomName, notify.RoomUrl))
	case ShowStatus_NoLiving:
		result = append(result, localutils.MessageTextf("斗鱼-%s直播结束了", notify.Nickname))
	}
	cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.GetAvatar().GetBig(), false, proxy_pool.PreferNone)
	if err != nil {
		logger.WithField("avatar", notify.GetAvatar().GetBig()).Errorf("upload avatar failed %v", err)
	} else {
		result = append(result, cover)
	}
	return result
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return logger.WithField("site", Site).
		WithFields(localutils.GroupLogFields(notify.GroupCode)).
		WithField("Name", notify.Nickname).
		WithField("Title", notify.RoomName).
		WithField("Status", notify.ShowStatus.String())
}

func NewConcernLiveNotify(groupCode int64, l *LiveInfo) *ConcernLiveNotify {
	if l == nil {
		return nil
	}
	return &ConcernLiveNotify{
		LiveInfo:  *l,
		GroupCode: groupCode,
	}
}

type Concern struct {
	*StateManager

	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
	stop      chan interface{}
	wg        sync.WaitGroup
}

func (c *Concern) Stop() {
	logger.Trace("正在停止douyu StateManager")
	c.StateManager.Stop()
	logger.Trace("douyu StateManager已停止")
	if c.stop != nil {
		close(c.stop)
	}
	close(c.eventChan)
	logger.Trace("正在停止douyu concern")
	c.wg.Wait()
	logger.Trace("douyu concern已停止")
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

	go c.EmitFreshCore("douyu", func(ctype concern.Type, id interface{}) error {
		roomid, ok := id.(int64)
		if !ok {
			return fmt.Errorf("cast fresh id type<%v> to int64 failed", reflect.ValueOf(id).Type().String())
		}
		if ctype.ContainAll(concern.DouyuLive) {
			oldInfo, _ := c.FindRoom(roomid, false)
			liveInfo, err := c.FindRoom(roomid, true)
			if err != nil {
				return fmt.Errorf("load liveinfo failed %v", err)
			}
			if oldInfo == nil {
				liveInfo.LiveStatusChanged = true
			}
			if oldInfo != nil && oldInfo.Living() != liveInfo.Living() {
				liveInfo.LiveStatusChanged = true
			}
			if oldInfo != nil && oldInfo.RoomName != liveInfo.RoomName {
				liveInfo.LiveTitleChanged = true
			}
			if oldInfo == nil || oldInfo.Living() != liveInfo.Living() || oldInfo.RoomName != liveInfo.RoomName {
				c.eventChan <- liveInfo
			}
		}
		return nil
	})
}

func (c *Concern) Add(groupCode int64, id int64, ctype concern.Type) (*LiveInfo, error) {
	var err error
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("id", id)

	err = c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}

	betardResp, err := Betard(id)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("查询房间信息失败 %v - %v", id, err)
	}
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	liveInfo := &LiveInfo{
		Nickname:   betardResp.GetRoom().GetNickname(),
		RoomId:     betardResp.GetRoom().GetRoomId(),
		RoomName:   betardResp.GetRoom().GetRoomName(),
		RoomUrl:    betardResp.GetRoom().GetRoomUrl(),
		ShowStatus: betardResp.GetRoom().GetShowStatus(),
		Avatar:     betardResp.GetRoom().GetAvatar(),
	}
	return liveInfo, nil
}

func (c *Concern) ListWatching(groupCode int64, p concern.Type) ([]*LiveInfo, []concern.Type, error) {
	log := logger.WithFields(localutils.GroupLogFields(groupCode))
	ids, ctypes, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(p)
	})
	if err != nil {
		return nil, nil, err
	}
	var resultTypes = make([]concern.Type, 0, len(ids))
	var result = make([]*LiveInfo, 0, len(ids))
	for index, id := range ids {
		liveInfo, err := c.FindOrLoadRoom(id.(int64))
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
	c.wg.Add(1)
	defer c.wg.Done()
	for ievent := range c.eventChan {
		switch ievent.Type() {
		case Live:
			event := ievent.(*LiveInfo)
			log := logger.WithField("name", event.GetNickname()).
				WithField("roomid", event.GetRoomId()).
				WithField("title", event.GetRoomName()).
				WithField("status", event.GetShowStatus().String()).
				WithField("videoLoop", event.GetVideoLoop().String())
			log.Debugf("debug event")

			groups, _, _, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern.Type) bool {
				return id.(int64) == event.RoomId && p.ContainAny(concern.DouyuLive)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}
			for _, groupCode := range groups {
				notify := NewConcernLiveNotify(groupCode, event)
				c.notify <- notify
				if event.Living() {
					log.Debug("living notify")
				} else {
					log.Debug("noliving notify")
				}
			}
		}
	}
}

func (c *Concern) FindRoom(id int64, load bool) (*LiveInfo, error) {
	var liveInfo *LiveInfo
	if load {
		betardResp, err := Betard(id)
		if err != nil {
			return nil, err
		}
		liveInfo = &LiveInfo{
			Nickname:   betardResp.GetRoom().GetNickname(),
			RoomId:     betardResp.GetRoom().GetRoomId(),
			RoomName:   betardResp.GetRoom().GetRoomName(),
			RoomUrl:    betardResp.GetRoom().GetRoomUrl(),
			ShowStatus: betardResp.GetRoom().GetShowStatus(),
			VideoLoop:  betardResp.GetRoom().GetVideoLoop(),
			Avatar:     betardResp.GetRoom().GetAvatar(),
		}
		_ = c.StateManager.AddLiveInfo(liveInfo)
	}
	if liveInfo != nil {
		return liveInfo, nil
	}
	return c.StateManager.GetLiveInfo(id)
}

func (c *Concern) FindOrLoadRoom(roomId int64) (*LiveInfo, error) {
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
