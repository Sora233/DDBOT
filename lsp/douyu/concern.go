package douyu

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/tidwall/buntdb"
)

var logger = utils.GetModuleLogger("douyu-concern")

type EventType int64

const (
	Live EventType = iota
)

type ConcernEvent interface {
	Type() EventType
}

func (m *LiveInfo) ToString() string {
	if m == nil {
		return ""
	}
	bin, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(bin)
}

func (m *LiveInfo) Living() bool {
	return m.ShowStatus == ShowStatus_Living && m.VideoLoop == VideoLoopStatus_Off
}

func (m *LiveInfo) Type() EventType {
	return Live
}

type ConcernLiveNotify struct {
	LiveInfo
	GroupCode int64 `json:"group_code"`
}

func (notify *ConcernLiveNotify) Type() concern.Type {
	return concern.DouyuLive
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

	emitChan  chan interface{}
	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
	stop      chan interface{}
	stopped   bool
}

func (c *Concern) Start() {
	db, err := localdb.GetClient()
	if err == nil {
		db.CreateIndex(c.GroupConcernStateKey(), c.GroupConcernStateKey("*"), buntdb.IndexString)
		db.CreateIndex(c.CurrentLiveKey(), c.CurrentLiveKey("*"), buntdb.IndexString)
		db.CreateIndex(c.FreshKey(), c.FreshKey("*"), buntdb.IndexString)
		db.CreateIndex(c.ConcernStateKey(), c.ConcernStateKey("*"), buntdb.IndexBinary)
	}

	err = c.StateManager.Start()
	if err != nil {
		logger.Errorf("state manager start err %v", err)
	}

	go c.notifyLoop()
	go c.emitFreshCore()
}

func (c *Concern) Add(groupCode int64, id int64, ctype concern.Type) (*LiveInfo, error) {
	var err error
	log := logger.WithField("GroupCode", groupCode)

	err = c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}

	betardResp, err := Betard(id)
	if err != nil {
		log.WithField("id", id).Error(err)
		return nil, fmt.Errorf("查询房间信息失败 %v - %v", id, err)
	}
	err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
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

func (c *Concern) ListLiving(groupCode int64, all bool) ([]*ConcernLiveNotify, error) {
	log := logger.WithField("group_code", groupCode).WithField("all", all)
	var result []*ConcernLiveNotify

	mids, _, err := c.StateManager.ListByGroup(groupCode, func(id int64, p concern.Type) bool {
		return p.ContainAny(concern.DouyuLive)
	})
	if err != nil {
		return nil, err
	}
	if len(mids) != 0 {
		result = make([]*ConcernLiveNotify, 0)
	}
	for _, mid := range mids {
		liveInfo, err := c.StateManager.GetLiveInfo(mid)
		if err != nil {
			log.WithField("mid", mid).Errorf("get LiveInfo err %v", err)
			continue
		}
		if all || liveInfo.Living() {
			result = append(result, NewConcernLiveNotify(groupCode, liveInfo))
		}
	}

	return result, nil
}

func (c *Concern) emitFreshCore() {
	for e := range c.emitChan {
		id, ok := e.(int64)
		if !ok {
			logger.WithField("emit", e).Errorf("emit element is not int64 id")
			continue
		}
		ctype, err := c.StateManager.GetConcern(id)
		if err != nil {
			logger.WithField("id", id).Errorf("get concern failed %v", err)
			continue
		}
		if ctype.ContainAll(concern.DouyuLive) {
			oldInfo, _ := c.findRoom(id, false)
			liveInfo, err := c.findRoom(id, true)
			if err != nil {
				logger.WithField("mid", id).Errorf("load liveinfo failed %v", err)
				continue
			}
			if oldInfo == nil || oldInfo.Living() != liveInfo.Living() || oldInfo.RoomName != liveInfo.RoomName {
				c.eventChan <- liveInfo
			}
		}
	}
}

func (c *Concern) notifyLoop() {
	for ievent := range c.eventChan {
		if c.stopped {
			return
		}
		switch ievent.Type() {
		case Live:
			event := ievent.(*LiveInfo)
			log := logger.WithField("name", event.GetNickname()).
				WithField("roomid", event.GetRoomId()).
				WithField("title", event.GetRoomName()).
				WithField("status", event.GetShowStatus().String()).
				WithField("videoLoop", event.GetVideoLoop().String())
			log.Debugf("debug event")

			groups, _, _, err := c.StateManager.List(func(groupCode int64, id int64, p concern.Type) bool {
				return id == event.RoomId && p.ContainAny(concern.DouyuLive)
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

func (c *Concern) findRoom(id int64, load bool) (*LiveInfo, error) {
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

func NewConcern(notify chan<- concern.Notify) *Concern {
	emitChan := make(chan interface{})
	c := &Concern{
		emitChan:     emitChan,
		StateManager: NewStateManager(emitChan),
		eventChan:    make(chan ConcernEvent, 500),
		notify:       notify,
		stop:         make(chan interface{}),
	}
	return c
}
