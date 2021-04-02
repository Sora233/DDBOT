package bilibili

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/Sora233/sliceutil"
	"github.com/tidwall/buntdb"
	"strconv"
	"sync"
	"time"
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
	go c.watchCore()
	go c.syncSub()
}

func (c *Concern) Stop() {
	c.stopped = true
	if c.stop != nil {
		close(c.stop)
	}
}

func (c *Concern) Add(groupCode int64, mid int64, ctype concern.Type) (*UserInfo, error) {
	var err error
	log := logger.WithField("GroupCode", groupCode).WithField("mid", mid)

	err = c.StateManager.CheckGroupConcern(groupCode, mid, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}

	infoResp, err := XSpaceAccInfo(mid)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v", mid, err)
	}
	if infoResp.Code != 0 {
		log.WithField("code", infoResp.Code).Errorf(infoResp.Message)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v %v", mid, infoResp.Code, infoResp.Message)
	}

	name := infoResp.GetData().GetName()

	if sliceutil.Contains([]int64{491474049}, mid) {
		return nil, fmt.Errorf("用户 %v 禁止watch", name)
	}

	oldCtype, err := c.StateManager.GetConcern(mid)
	if err != nil {
		log.Errorf("get concern error %v", err)
	} else if oldCtype.Empty() {
		err = c.modifyUserRelation(mid, ActSub)
		if err != nil {
			return nil, fmt.Errorf("关注用户失败 - 内部错误")
		}
	}

	_, err = c.StateManager.AddGroupConcern(groupCode, mid, ctype)
	if err != nil {
		log.Errorf("AddGroupConcern error %v", err)
		return nil, fmt.Errorf("关注用户失败 - 内部错误")
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

func (c *Concern) Remove(groupCode int64, mid int64, ctype concern.Type) (concern.Type, error) {
	newCtype, err := c.StateManager.RemoveGroupConcern(groupCode, mid, ctype)
	if err != nil {
		return concern.Empty, err
	}

	{
		// inner err is not outer err
		state, err := c.GetConcern(mid)
		if err != nil {
			logger.WithField("mid", mid).Errorf("GetConcern error %v", err)
		} else if state.Empty() {
			logger.WithField("mid", mid).Debug("empty state, unsub")
			c.modifyUserRelation(mid, ActUnsub)
		}
	}
	return newCtype, err
}

func (c *Concern) ListWatching(groupCode int64, ctype concern.Type) ([]*UserInfo, error) {
	log := logger.WithField("group_code", groupCode)

	mids, _, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(ctype)
	})
	if err != nil {
		return nil, err
	}
	var result = make([]*UserInfo, 0)
	for _, mid := range mids {
		userInfo, err := c.StateManager.GetUserInfo(mid.(int64))
		if err != nil {
			log.WithField("mid", mid).Errorf("GetUserInfo error %v", err)
			continue
		}
		result = append(result, userInfo)
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
				log.WithField("group_code", groupCode).Debug("news notify")
				notify := NewConcernNewsNotify(groupCode, event)
				c.notify <- notify
			}
		}

	}
}

func (c *Concern) watchCore() {
	t := time.NewTimer(time.Second * 30)
	var wg sync.WaitGroup
	for {
		<-t.C
		if c.stopped {
			return
		}
		start := time.Now()

		wg.Add(2)
		go func() {
			defer wg.Done()
			defer logger.Tracef("watchCore dynamic fresh done")
			newsList, err := c.freshDynamicNew()
			if err != nil {
				logger.Errorf("freshDynamicNew failed %v", err)
				return
			} else {
				for _, news := range newsList {
					c.eventChan <- news
				}
			}
		}()

		go func() {
			defer wg.Done()
			defer logger.Tracef("watchCore live fresh done")
			liveInfo, err := c.freshLive()
			if err != nil {
				logger.Errorf("freshLive error %v", err)
				return
			}
			var liveInfoMap = make(map[int64]*LiveInfo)
			for _, info := range liveInfo {
				liveInfoMap[info.Mid] = info
			}

			_, ids, types, err := c.List(func(groupCode int64, id interface{}, p concern.Type) bool {
				return p.ContainAny(concern.BibiliLive)
			})
			if err != nil {
				logger.Errorf("List error %v", err)
				return
			}
			ids, types, err = c.GroupTypeById(ids, types)
			if err != nil {
				logger.Errorf("GroupTypeById error %v", err)
				return
			}

			for _, id := range ids {
				mid := id.(int64)
				oldInfo, err := c.GetLiveInfo(mid)
				if err == buntdb.ErrNotFound || oldInfo == nil {
					// first live info
					if newInfo, found := liveInfoMap[mid]; found {
						c.eventChan <- newInfo
						c.AddLiveInfo(newInfo)
					}
					continue
				}
				if oldInfo.Status == LiveStatus_NoLiving {
					if newInfo, found := liveInfoMap[mid]; found {
						// notliving -> living
						c.eventChan <- newInfo
						c.AddLiveInfo(newInfo)
					}
				} else if oldInfo.Status == LiveStatus_Living {
					if newInfo, found := liveInfoMap[mid]; !found {
						// living -> notliving
						newInfo = NewLiveInfo(&oldInfo.UserInfo, oldInfo.LiveTitle, oldInfo.Cover, LiveStatus_NoLiving)
						c.eventChan <- newInfo
						c.AddLiveInfo(newInfo)
					} else {
						if newInfo.LiveTitle != oldInfo.LiveTitle {
							// live title change
							c.eventChan <- newInfo
							c.AddLiveInfo(newInfo)
						}
					}
				}
			}
		}()
		wg.Wait()
		t.Reset(time.Second * 30)
		end := time.Now()
		logger.WithField("cost", end.Sub(start)).Debug("watchCore loop done")
	}
}

func (c *Concern) freshDynamicNew() ([]*NewsInfo, error) {
	resp, err := DynamicSrvDynamicNew()
	if err != nil {
		return nil, err
	}
	var newsMap = make(map[int64][]*Card)
	if resp.GetCode() != 0 {
		logger.WithField("code", resp.GetCode()).
			WithField("msg", resp.GetMessage()).
			Errorf("fresh dynamic new failed")
		return nil, errors.New(resp.Message)
	}
	now := time.Now()
	for _, card := range resp.GetData().GetCards() {
		uid := card.GetDesc().GetUid()
		t := time.Unix(int64(card.GetDesc().GetTimestamp()), 0)
		replaced, err := c.MarkDynamicId(card.GetDesc().GetDynamicId())
		if err != nil || replaced {
			continue
		}
		if t.Add(time.Second * 90).Before(now) {
			continue
		}
		newsMap[uid] = append(newsMap[uid], card)
	}
	var result []*NewsInfo
	for uid, cards := range newsMap {
		userInfo, err := c.StateManager.GetUserInfo(uid)
		if err == buntdb.ErrNotFound {
			continue
		} else if err != nil {
			logger.WithField("mid", uid).Debugf("find user info error %v", err)
			continue
		}
		result = append(result, NewNewsInfoWithDetail(userInfo, cards))
	}
	logger.WithField("NewsInfo Size", len(result)).Tracef("freshDynamicNew done")
	return result, nil
}

// return all LiveInfo in LiveStatus_Living
func (c *Concern) freshLive() ([]*LiveInfo, error) {
	var liveInfo []*LiveInfo
	var infoSet = make(map[int64]bool)
	for {
		resp, err := FeedList()
		if err != nil {
			logger.Errorf("freshLive FeedList error %v", err)
			return nil, err
		} else if resp.GetCode() != 0 {
			logger.Errorf("freshLive FeedList code %v msg %v", resp.GetCode(), resp.GetMessage())
			return nil, err
		}
		pageSize, _ := strconv.ParseInt(resp.GetData().GetPagesize(), 10, 64)
		for _, l := range resp.GetData().GetList() {
			if infoSet[l.GetUid()] {
				continue
			}
			infoSet[l.GetUid()] = true
			liveInfo = append(liveInfo, NewLiveInfo(
				NewUserInfo(l.GetUid(), l.GetRoomid(), l.GetUname(), l.GetLink()),
				l.GetTitle(),
				l.GetPic(),
				LiveStatus_Living,
			))
		}
		if (int(pageSize) != 0 && len(resp.GetData().GetList()) != int(pageSize)) || len(resp.GetData().GetList()) == 0 {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	logger.WithField("LiveInfo Size", len(liveInfo)).Tracef("freshLive done")
	return liveInfo, nil
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

func (c *Concern) modifyUserRelation(mid int64, act int) error {
	resp, err := RelationModify(mid, act)
	if err != nil {
		return err
	}
	if resp.GetCode() != 0 {
		logger.WithField("code", resp.GetCode()).
			WithField("message", resp.GetMessage()).
			WithField("act", act).
			WithField("mid", mid).
			Errorf("modifyUserRelation error")
		return fmt.Errorf("%v %v", resp.GetCode(), resp.GetMessage())
	}
	logger.WithField("mid", mid).WithField("act", act).Debug("modify relation")
	return nil
}

func (c *Concern) syncSub() {
	defer logger.Debug("syncSub done")
	resp, err := GetAttentionList()
	if err != nil {
		logger.Errorf("syncSub error %v", err)
		return
	}
	if resp.GetCode() != 0 {
		logger.WithField("code", resp.GetCode()).
			WithField("msg", resp.GetMessage()).
			Errorf("syncSub GetAttentionList error")
		return
	}
	var midSet = make(map[int64]bool)
	var attentionMidSet = make(map[int64]bool)
	_, _, _, err = c.List(func(groupCode int64, id interface{}, p concern.Type) bool {
		midSet[id.(int64)] = true
		return true
	})
	if err != nil {
		logger.Errorf("syncSub List all error %v", err)
		return
	}
	for _, attentionMid := range resp.GetData().GetList() {
		attentionMidSet[attentionMid] = true
	}
	for mid := range midSet {
		if _, found := attentionMidSet[mid]; !found {
			c.modifyUserRelation(mid, ActSub)
			time.Sleep(time.Second * 3)
		}
	}
}

func (c *Concern) findOrLoadUser(mid int64) (*UserInfo, error) {
	info, err := c.findUser(mid, false)
	if err != nil || info == nil {
		info, err = c.findUser(mid, true)
	}
	return info, err
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
