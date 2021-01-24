package bilibili

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/Sora233/sliceutil"
)

var logger = utils.GetModuleLogger("bilibili-concern")

type EventType int64

const (
	Live EventType = iota
	News
)

type ConcernEvent interface {
	Type() EventType
}

type Concern struct {
	*StateManager

	eventChan chan ConcernEvent
	emitChan  chan interface{}
	notify    chan<- concern.Notify
	stopped   bool
	stop      chan interface{}
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

func (c *Concern) Start() {
	err := c.StateManager.Start()
	if err != nil {
		logger.Errorf("state manager start err %v", err)
	}

	go c.notifyLoop()
	go c.EmitFreshCore("bilibili", func(ctype concern.Type, id interface{}) error {
		mid, ok := id.(int64)
		if !ok {
			return errors.New("cast fresh id to int64 failed")
		}
		if ctype.ContainAll(concern.BibiliLive) {
			c.freshLive(mid)
		}
		if ctype.ContainAll(concern.BilibiliNews) {
			c.freshNews(mid)
		}
		return nil
	})
}

func (c *Concern) Stop() {
	c.stopped = true
	if c.stop != nil {
		close(c.stop)
	}
}

func (c *Concern) Add(groupCode int64, mid int64, ctype concern.Type) (*UserInfo, error) {
	var err error
	log := logger.WithField("GroupCode", groupCode)

	err = c.StateManager.CheckGroupConcern(groupCode, mid, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}

	infoResp, err := XSpaceAccInfo(mid)
	if err != nil {
		log.WithField("mid", mid).Error(err)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v", mid, err)
	}
	if infoResp.Code != 0 {
		log.WithField("mid", mid).WithField("code", infoResp.Code).Errorf(infoResp.Message)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v %v", mid, infoResp.Code, infoResp.Message)
	}

	name := infoResp.GetData().GetName()

	if sliceutil.Contains([]int64{491474049}, mid) {
		return nil, fmt.Errorf("用户 %v 禁止watch", name)
	}
	if ctype.ContainAll(concern.BibiliLive) {
		if RoomStatus(infoResp.GetData().GetLiveRoom().GetRoomStatus()) == RoomStatus_NonExist {
			return nil, fmt.Errorf("用户 %v 暂未开通直播间", name)
		}
	}

	err = c.StateManager.AddGroupConcern(groupCode, mid, ctype)
	if err != nil {
		return nil, err
	}

	userInfo := NewUserInfo(
		mid,
		infoResp.GetData().GetLiveRoom().GetRoomid(),
		infoResp.GetData().GetName(),
		infoResp.GetData().GetLiveRoom().GetUrl(),
	)

	_ = c.StateManager.AddUserInfo(userInfo)
	return userInfo, nil
}

func (c *Concern) ListLiving(groupCode int64, all bool) ([]*ConcernLiveNotify, error) {
	log := logger.WithField("group_code", groupCode).WithField("all", all)
	var result []*ConcernLiveNotify

	mids, _, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(concern.BibiliLive)
	})
	if err != nil {
		return nil, err
	}
	if len(mids) != 0 {
		result = make([]*ConcernLiveNotify, 0)
	}
	for _, mid := range mids {
		liveInfo, err := c.StateManager.GetLiveInfo(mid.(int64))
		if err != nil {
			log.WithField("mid", mid).Errorf("get LiveInfo err %v", err)
			continue
		}
		if all || liveInfo.Status == LiveStatus_Living {
			result = append(result, NewConcernLiveNotify(groupCode, liveInfo))
		}
	}
	return result, nil
}

func (c *Concern) ListNews(groupCode int64, all bool) ([]*ConcernNewsNotify, error) {
	log := logger.WithField("group_code", groupCode).WithField("all", all)
	var result []*ConcernNewsNotify

	mids, _, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(concern.BilibiliNews)
	})
	if err != nil {
		return nil, err
	}
	if len(mids) != 0 {
		result = make([]*ConcernNewsNotify, 0)
	}
	for _, mid := range mids {
		newsInfo, err := c.StateManager.GetNewsInfo(mid.(int64))
		if err != nil {
			log.WithField("mid", mid).Errorf("get newsInfo err %v", err)
			continue
		}
		result = append(result, NewConcernNewsNotify(groupCode, newsInfo))
	}
	return result, nil
}

func (c *Concern) notifyLoop() {
	for ievent := range c.eventChan {
		if c.stopped {
			return
		}

		switch ievent.Type() {
		case Live:
			event := (ievent).(*LiveInfo)
			log := logger.WithField("mid", event.Mid).
				WithField("name", event.Name).
				WithField("roomid", event.RoomId).
				WithField("title", event.LiveTitle).
				WithField("status", event.Status.String()).
				WithField("type", event.Type())
			log.Debugf("debug event")

			groups, _, _, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern.Type) bool {
				return id.(int64) == event.Mid && p.ContainAny(concern.BibiliLive)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}

			for _, groupCode := range groups {
				notify := NewConcernLiveNotify(groupCode, event)
				c.notify <- notify
				if event.Status == LiveStatus_Living {
					log.WithField("group_code", groupCode).Debug("living notify")
				} else if event.Status == LiveStatus_NoLiving {
					log.WithField("group_code", groupCode).Debug("noliving notify")
				} else {
					log.WithField("group_code", groupCode).Error("unknown live status")
				}
			}
		case News:
			event := (ievent).(*NewsInfo)
			log := logger.WithField("mid", event.Mid).
				WithField("name", event.Name).
				WithField("news_number", len(event.Cards)).
				WithField("type", event.Type())
			log.Debugf("debug event")

			groups, _, _, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern.Type) bool {
				return id.(int64) == event.Mid && p.ContainAny(concern.BilibiliNews)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}
			for _, groupCode := range groups {
				notify := NewConcernNewsNotify(groupCode, event)
				c.notify <- notify
			}
		}

	}
}

func (c *Concern) freshLive(mid int64) {
	oldLiveInfo, _ := c.findUserLiving(mid, false)
	newLiveInfo, err := c.findUserLiving(mid, true)
	if err != nil {
		logger.WithField("mid", mid).Errorf("load liveinfo failed %v", err)
		return
	}
	if oldLiveInfo == nil || oldLiveInfo.Status != newLiveInfo.Status || oldLiveInfo.LiveTitle != newLiveInfo.LiveTitle {
		c.eventChan <- newLiveInfo
	}
}

func (c *Concern) freshNews(mid int64) {
	oldNewsInfo, _ := c.findUserNews(mid, false)
	newNewsInfo, err := c.findUserNews(mid, true)
	if err != nil {
		logger.WithField("mid", mid).Errorf("load newsinfo failed %v", err)
		return
	}
	if oldNewsInfo == nil {
		logger.WithField("mid", mid).Debugf("oldNewsInfo nil, skip notify")
		return
	}
	if oldNewsInfo.Timestamp > newNewsInfo.Timestamp {
		logger.WithField("mid", mid).
			WithField("old_timestamp", oldNewsInfo.Timestamp).
			WithField("new_timestamp", newNewsInfo.Timestamp).
			Debugf("newNewsInfo timestamp is less than oldNewsInfo timestamp, " +
				"maybe some dynamic is deleted, clear newsInfo.")
		err := c.clearUserNews(mid)
		if err != nil {
			logger.WithField("mid", err).Errorf("clear user NewsInfo err %v", err)
		}
		return
	}
	if newNewsInfo.LastDynamicId == 0 || len(newNewsInfo.Cards) == 0 {
		logger.WithField("mid", mid).Debugf("newNewsInfo is empty")
		return
	}
	if oldNewsInfo.LastDynamicId != newNewsInfo.LastDynamicId {
		var newIndex = 1 // if too many news, just notify latest one
		for index := range newNewsInfo.Cards {
			if oldNewsInfo.LastDynamicId == newNewsInfo.Cards[index].GetDesc().GetDynamicId() {
				newIndex = index
			}
		}
		newNewsInfo.Cards = newNewsInfo.Cards[:newIndex]
		if len(newNewsInfo.Cards) > 0 {
			c.eventChan <- newNewsInfo
		}
	}
}

func (c *Concern) findUser(mid int64, load bool) (*UserInfo, error) {
	if load {
		resp, err := XSpaceAccInfo(mid)
		if err != nil {
			return nil, err
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("code:%v %v", resp.Code, resp.Message)
		}
		newUserInfo := NewUserInfo(mid,
			resp.GetData().GetLiveRoom().GetRoomid(),
			resp.GetData().GetName(),
			resp.GetData().GetLiveRoom().GetUrl(),
		)
		err = c.StateManager.AddUserInfo(newUserInfo)
		if err != nil {
			return nil, err
		}
	}
	return c.StateManager.GetUserInfo(mid)
}

func (c *Concern) findUserLiving(mid int64, load bool) (*LiveInfo, error) {
	userInfo, err := c.findUser(mid, load)
	if err != nil {
		return nil, err
	}

	var liveInfo *LiveInfo

	if load {
		roomInfo, err := GetRoomInfoOld(userInfo.Mid)
		if err != nil {
			return nil, err
		}
		if roomInfo.Code != 0 {
			return nil, fmt.Errorf("code:%v %v", roomInfo.Code, roomInfo.Message)
		}
		liveInfo = NewLiveInfo(userInfo,
			roomInfo.GetData().GetTitle(),
			roomInfo.GetData().GetCover(),
			roomInfo.GetData().GetLiveStatus(),
		)
		_ = c.StateManager.AddLiveInfo(liveInfo)
	}

	if liveInfo != nil {
		return liveInfo, nil
	}
	return c.StateManager.GetLiveInfo(mid)
}

func (c *Concern) findUserNews(mid int64, load bool) (*NewsInfo, error) {
	userInfo, err := c.findUser(mid, load)
	if err != nil {
		return nil, err
	}

	var newsInfo *NewsInfo

	if load {
		history, err := DynamicSrvSpaceHistory(mid)
		if err != nil {
			return nil, err
		}
		if history.Code != 0 {
			return nil, fmt.Errorf("code:%v %v", history.Code, history.Message)
		}
		newsInfo = NewNewsInfoWithDetail(userInfo, history.GetData().GetCards())
		_ = c.StateManager.AddNewsInfo(newsInfo)
	}
	if newsInfo != nil {
		return newsInfo, nil
	}
	return c.StateManager.GetNewsInfo(mid)
}

func (c *Concern) clearUserNews(mid int64) error {
	newsInfo, err := c.findUserNews(mid, false)
	if err != nil {
		return err
	}
	return c.StateManager.DeleteNewsInfo(newsInfo)
}
