package douyu

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/msg"
	localutils "github.com/Sora233/DDBOT/utils"
	"reflect"
	"runtime"
	"sync"
)

var logger = utils.GetModuleLogger("douyu-concern")

const (
	Live concern_type.Type = "live"
)

type Concern struct {
	*StateManager

	eventChan chan concernEvent
	notify    chan<- concern.Notify
	stop      chan interface{}
	wg        sync.WaitGroup
}

func (c *Concern) Site() string {
	return "douyu"
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return ParseUid(s)
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
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

func (c *Concern) Start() error {
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

	go c.EmitFreshCore(Site, func(ctype concern_type.Type, id interface{}) error {
		roomid, ok := id.(int64)
		if !ok {
			return fmt.Errorf("cast fresh id type<%v> to int64 failed", reflect.ValueOf(id).Type().String())
		}
		if ctype.ContainAll(Live) {
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
	return nil
}

func (c *Concern) Add(ctx msg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(int64)
	var err error
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("id", id)

	err = c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		if err == concern.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}
	liveInfo, err := c.FindOrLoadRoom(id)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("查询房间信息失败 %v - %v", id, err)
	}
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(id, liveInfo.Nickname), nil
}

func (c *Concern) Remove(ctx msg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(int64)
	identity, _ := c.Get(id)
	_, err := c.StateManager.RemoveGroupConcern(groupCode, id, ctype)
	return identity, err
}

func (c *Concern) List(groupCode int64, p concern_type.Type) ([]concern.IdentityInfo, []concern_type.Type, error) {
	log := logger.WithFields(localutils.GroupLogFields(groupCode))
	_, ids, ctypes, err := c.StateManager.List(func(_groupCode int64, id interface{}, p concern_type.Type) bool {
		return _groupCode == groupCode && p.ContainAny(p)
	})
	if err != nil {
		return nil, nil, err
	}
	ids, ctypes, err = c.StateManager.GroupTypeById(ids, ctypes)
	if err != nil {
		return nil, nil, err
	}
	var result = make([]concern.IdentityInfo, 0, len(ids))
	var resultTypes = make([]concern_type.Type, 0, len(ids))
	for index, id := range ids {
		liveInfo, err := c.FindOrLoadRoom(id.(int64))
		if err != nil {
			log.WithField("id", id).Errorf("get LiveInfo err %v", err)
			continue
		}
		result = append(result, concern.NewIdentity(liveInfo.GetRoomId(), liveInfo.GetNickname()))
		resultTypes = append(resultTypes, ctypes[index])
	}
	return result, resultTypes, nil
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	liveInfo, err := c.FindOrLoadRoom(id.(int64))
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(liveInfo.GetRoomId(), liveInfo.GetNickname()), nil
}

func (c *Concern) notifyLoop() {
	c.wg.Add(1)
	defer c.wg.Done()
	for ievent := range c.eventChan {
		switch ievent.Type() {
		case Live:
			event := ievent.(*LiveInfo)
			log := event.Logger()
			log.Debugf("debug event")

			groups, _, _, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern_type.Type) bool {
				return id.(int64) == event.RoomId && p.ContainAny(Live)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}
			for _, groupCode := range groups {
				notify := NewConcernLiveNotify(groupCode, event)
				c.notify <- notify
				if event.Living() {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debug("living notify")
				} else {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debug("noliving notify")
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
		eventChan:    make(chan concernEvent, 500),
		notify:       notify,
		stop:         make(chan interface{}),
	}
	return c
}
