package bilibili

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/concern"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"golang.org/x/sync/errgroup"
	"runtime"
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
	stop      chan interface{}
	wg        sync.WaitGroup
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

func (c *Concern) Stop() {
	logger.Trace("正在停止bilibili StateManager")
	c.StateManager.Stop()
	logger.Trace("bilibili StateManager已停止")
	if c.stop != nil {
		close(c.stop)
	}
	logger.Trace("正在停止bilibili concern")
	c.wg.Wait()
	logger.Trace("bilibili concern已停止")
}

func (c *Concern) Start() {
	err := c.StateManager.Start()
	if err != nil {
		logger.Errorf("state manager start err %v", err)
	}

	if !IsVerifyGiven() {
		logger.Errorf("注意：B站配置不完整，B站相关功能无法使用！")
		return
	}

	if runtime.NumCPU() >= 3 {
		for i := 0; i < 3; i++ {
			go c.notifyLoop()
		}
	} else {
		go c.notifyLoop()
	}

	go c.watchCore()
	go func() {
		c.SyncSub()

		tick := time.Tick(time.Hour)
		for {
			select {
			case <-tick:
				c.SyncSub()
			case <-c.stop:
				return
			}
		}
	}()
}

func (c *Concern) Add(groupCode int64, mid int64, ctype concern.Type) (*UserInfo, error) {
	var err error
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("mid", mid)

	err = c.StateManager.CheckGroupConcern(groupCode, mid, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		log.Errorf("CheckGroupConcern error %v", err)
		return nil, err
	}
	var userInfo *UserInfo

	userInfo, _ = c.GetUserInfo(mid)
	if userInfo == nil {
		infoResp, err := XSpaceAccInfo(mid)
		if err != nil {
			log.Errorf("XSpaceAccInfo error %v", err)
			return nil, fmt.Errorf("查询用户信息失败 %v - %v", mid, err)
		}
		if infoResp.Code != 0 {
			log.WithField("code", infoResp.Code).Errorf(infoResp.Message)
			return nil, fmt.Errorf("查询用户信息失败 %v - %v %v", mid, infoResp.Code, infoResp.Message)
		}

		userInfo = NewUserInfo(
			mid,
			infoResp.GetData().GetLiveRoom().GetRoomid(),
			infoResp.GetData().GetName(),
			infoResp.GetData().GetLiveRoom().GetUrl(),
		)
		log = log.WithField("name", userInfo.GetName())
	} else {
		log = log.WithField("name", userInfo.GetName())
		log.Debugf("UserInfo cache hit")
	}

	userStat, err := c.StatUserWithCache(mid, time.Hour)
	if err != nil {
		log.Errorf("get UserStat error %v\n", err)
	} else if userStat != nil {
		userInfo.UserStat = userStat
	}

	oldCtype, err := c.StateManager.GetConcern(mid)
	if err != nil {
		log.Errorf("GetConcern error %v", err)
	} else if oldCtype.Empty() {
		resp, err := c.ModifyUserRelation(mid, ActSub)
		if err != nil {
			if err == ErrVerifyRequired {
				log.Errorf("ModifyUserRelation error %v", err)
				return nil, fmt.Errorf("关注用户失败 - 未配置B站")
			} else {
				log.WithField("action", ActSub).Errorf("ModifyUserRelation error %v", err)
				return nil, fmt.Errorf("关注用户失败 - 内部错误")
			}
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("关注用户失败 - %v", resp.GetMessage())
		}
	}

	_, err = c.StateManager.AddGroupConcern(groupCode, mid, ctype)
	if err != nil {
		log.Errorf("AddGroupConcern error %v", err)
		return nil, fmt.Errorf("关注用户失败 - 内部错误")
	}
	err = c.StateManager.SetUidFirstTimestampIfNotExist(mid, time.Now().Add(-time.Second*30).Unix())
	if err != nil {
		log.Errorf("SetUidFirstTimestampIfNotExist failed %v", err)
	}

	_ = c.StateManager.AddUserInfo(userInfo)
	return userInfo, nil
}

func (c *Concern) Remove(groupCode int64, mid int64, ctype concern.Type) (concern.Type, error) {
	var newCtype concern.Type
	err := c.StateManager.RWCoverTx(func(tx *buntdb.Tx) error {
		var (
			err      error
			allCtype concern.Type
		)
		newCtype, err = c.StateManager.RemoveGroupConcern(groupCode, mid, ctype)
		if err != nil {
			return err
		}
		if !ctype.ContainAll(concern.BibiliLive) {
			return nil
		}
		allCtype, err = c.StateManager.ListById(mid)
		if err != nil {
			return err
		}
		// 如果已经没有watch live的了，此时应该把liveinfo删掉，否则会无法刷新到livelinfo
		// 如果此时liveinfo是living状态，则此状态会一直保留，下次watch时会以为在living错误推送
		if !allCtype.ContainAll(concern.BibiliLive) {
			err = c.StateManager.DeleteLiveInfo(mid)
			if err == buntdb.ErrNotFound {
				err = nil
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	return newCtype, err
}

func (c *Concern) ListWatching(groupCode int64, ctype concern.Type) ([]*UserInfo, []concern.Type, error) {
	log := logger.WithFields(localutils.GroupLogFields(groupCode))

	mids, ctypes, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(ctype)
	})
	if err != nil {
		return nil, nil, err
	}
	var result = make([]*UserInfo, 0, len(mids))
	var resultTypes = make([]concern.Type, 0, len(mids))
	for index, mid := range mids {
		userInfo, err := c.StateManager.GetUserInfo(mid.(int64))
		if err != nil {
			log.WithField("mid", mid).Errorf("GetUserInfo error %v", err)
			continue
		}
		result = append(result, userInfo)
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
			event := (ievent).(*LiveInfo)
			log := event.Logger()
			log.Debugf("new event - live notify")

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
					log.WithFields(localutils.GroupLogFields(groupCode)).Debug("living notify")
				} else if event.Status == LiveStatus_NoLiving {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debug("noliving notify")
				} else {
					log.WithFields(localutils.GroupLogFields(groupCode)).Error("unknown live status")
				}
			}
		case News:
			event := (ievent).(*NewsInfo)
			log := event.Logger()
			log.Debugf("new event - news notify")

			groups, _, _, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern.Type) bool {
				return id.(int64) == event.Mid && p.ContainAny(concern.BilibiliNews)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}
			for _, groupCode := range groups {
				log.WithFields(localutils.GroupLogFields(groupCode)).Debug("news notify")
				notifies := NewConcernNewsNotify(groupCode, event)
				for _, notify := range notifies {
					c.notify <- notify
				}
			}
		}

	}
}

func (c *Concern) watchCore() {
	c.wg.Add(1)
	defer c.wg.Done()
	t := time.NewTimer(time.Second * 3)
	for {
		select {
		case <-t.C:
		case <-c.stop:
			close(c.eventChan)
			return
		}
		start := time.Now()
		var errGroup errgroup.Group

		errGroup.Go(func() error {
			defer func() { logger.WithField("cost", time.Now().Sub(start)).Tracef("watchCore dynamic fresh done") }()
			newsList, err := c.freshDynamicNew()
			if err != nil {
				logger.Errorf("freshDynamicNew failed %v", err)
				return err
			} else {
				for _, news := range newsList {
					c.eventChan <- news
				}
			}
			return nil
		})

		errGroup.Go(func() error {
			defer func() { logger.WithField("cost", time.Now().Sub(start)).Tracef("watchCore live fresh done") }()
			liveInfo, err := c.freshLive()
			if err != nil {
				logger.Errorf("freshLive error %v", err)
				return err
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
				return err
			}
			ids, types, err = c.GroupTypeById(ids, types)
			if err != nil {
				logger.Errorf("GroupTypeById error %v", err)
				return err
			}

			sendLiveInfo := func(info *LiveInfo) {
				addLiveInfoErr := c.AddLiveInfo(info)
				if addLiveInfoErr != nil {
					// 如果因为系统原因add失败，会造成重复推送
					// 按照ddbot的原则，选择不推送，而非重复推送
					logger.WithField("mid", info.Mid).Errorf("add live info error %v", err)
				} else {
					c.eventChan <- info
				}
			}

			for _, id := range ids {
				mid := id.(int64)
				oldInfo, _ := c.GetLiveInfo(mid)
				if oldInfo == nil {
					// first live info
					if newInfo, found := liveInfoMap[mid]; found {
						newInfo.LiveStatusChanged = true
						sendLiveInfo(newInfo)
					}
					continue
				}
				if oldInfo.Status == LiveStatus_NoLiving {
					if newInfo, found := liveInfoMap[mid]; found {
						// notliving -> living
						newInfo.LiveStatusChanged = true
						sendLiveInfo(newInfo)
					}
				} else if oldInfo.Status == LiveStatus_Living {
					if newInfo, found := liveInfoMap[mid]; !found {
						// living -> notliving
						if count := c.IncNotLiveCount(mid); count < 3 {
							logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
								WithField("notlive_count", count).
								Debug("notlive counting")
							continue
						} else {
							logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
								Debug("notlive count done, notlive confirmed")
						}
						c.ClearNotLiveCount(mid)
						newInfo = NewLiveInfo(&oldInfo.UserInfo, oldInfo.LiveTitle, oldInfo.Cover, LiveStatus_NoLiving)
						newInfo.LiveStatusChanged = true
						sendLiveInfo(newInfo)
					} else {
						if newInfo.LiveTitle == "bilibili主播的直播间" {
							newInfo.LiveTitle = oldInfo.LiveTitle
						}
						c.ClearNotLiveCount(mid)
						if newInfo.LiveTitle != oldInfo.LiveTitle {
							// live title change
							newInfo.LiveTitleChanged = true
							sendLiveInfo(newInfo)
						}
					}
				}
			}
			return nil
		})
		err := errGroup.Wait()
		end := time.Now()
		if err == nil {
			logger.WithField("cost", end.Sub(start)).Tracef("watchCore loop done")
		} else {
			logger.WithField("cost", end.Sub(start)).Errorf("watchCore error %v", err)
		}
		t.Reset(config.GlobalConfig.GetDuration("bilibili.interval"))
	}
}

func (c *Concern) freshDynamicNew() ([]*NewsInfo, error) {
	var start = time.Now()
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
	logger.WithField("cost", time.Now().Sub(start)).Trace("freshDynamicNew cost 1")
	for _, card := range resp.GetData().GetCards() {
		uid := card.GetDesc().GetUid()
		// 应该用dynamic_id_str
		// 但好像已经没法保持向后兼容同时改动了
		// 只能相信概率论了，出问题的概率应该比较小，出问题会导致推送丢失
		replaced, err := c.MarkDynamicId(card.GetDesc().GetDynamicId())
		if err != nil || replaced {
			if err != nil {
				logger.WithField("uid", uid).
					WithField("dynamicId", card.GetDesc().GetDynamicId()).
					Errorf("MarkDynamicId error %v", err)
			}
			continue
		}
		ts, err := c.StateManager.GetUidFirstTimestamp(uid)
		if err == nil && card.GetDesc().GetTimestamp() < ts {
			logger.WithField("uid", uid).
				WithField("dynamicId", card.GetDesc().GetDynamicId()).
				Debugf("past news skip")
			continue
		}
		newsMap[uid] = append(newsMap[uid], card)
	}
	logger.WithField("cost", time.Now().Sub(start)).Trace("freshDynamicNew cost 2")
	var result []*NewsInfo
	for uid, cards := range newsMap {
		userInfo, err := c.StateManager.GetUserInfo(uid)
		if err == buntdb.ErrNotFound {
			continue
		} else if err != nil {
			logger.WithField("mid", uid).Debugf("find user info error %v", err)
			continue
		}
		if len(cards) > 0 {
			// 如果更新了名字，有机会在这里捞回来
			userInfo.Name = cards[0].GetDesc().GetUserProfile().GetInfo().GetUname()
		}
		if len(cards) > 3 {
			// 有时候b站抽风会刷屏
			cards = cards[:3]
		}
		result = append(result, NewNewsInfoWithDetail(userInfo, cards))
	}
	logger.WithField("cost", time.Now().Sub(start)).WithField("NewsInfo Size", len(result)).Trace("freshDynamicNew done")
	return result, nil
}

// return all LiveInfo in LiveStatus_Living
func (c *Concern) freshLive() ([]*LiveInfo, error) {
	var start = time.Now()
	var liveInfo []*LiveInfo
	var infoSet = make(map[int64]bool)
	var page = 1
	var maxPage int32 = 1
	var zeroCount = 0
	for {
		resp, err := FeedList(FeedPageOpt(page))
		if err != nil {
			logger.Errorf("freshLive FeedList error %v", err)
			return nil, err
		} else if resp.GetCode() != 0 {
			logger.Errorf("freshLive FeedList code %v msg %v", resp.GetCode(), resp.GetMessage())
			return nil, err
		}
		var (
			dataSize    = len(resp.GetData().GetList())
			pageSize, _ = strconv.ParseInt(resp.GetData().GetPagesize(), 10, 32)
			curTotal    = resp.GetData().GetResults()
			curMaxPage  = (curTotal-1)/int32(pageSize) + 1
		)
		logger.WithFields(logrus.Fields{
			"CurTotal":   curTotal,
			"PageSize":   pageSize,
			"CurMaxPage": curMaxPage,
			"maxPage":    maxPage,
			"page":       page,
		}).Trace("freshLive debug")
		if curMaxPage > maxPage {
			maxPage = curMaxPage
		}
		for _, l := range resp.GetData().GetList() {
			if infoSet[l.GetUid()] {
				continue
			}
			infoSet[l.GetUid()] = true
			info := NewLiveInfo(
				NewUserInfo(l.GetUid(), l.GetRoomid(), l.GetUname(), l.GetLink()),
				l.GetTitle(),
				l.GetPic(),
				LiveStatus_Living,
			)
			if info.Cover == "" {
				info.Cover = l.GetCover()
			}
			if info.Cover == "" {
				info.Cover = l.GetFace()
			}
			liveInfo = append(liveInfo, info)
		}
		if dataSize != 0 {
			zeroCount = 0
			page++
		} else {
			zeroCount += 1
		}
		if int32(page) > maxPage || zeroCount >= 3 {
			break
		}
	}
	logger.WithField("cost", time.Now().Sub(start)).WithField("Page", page).WithField("LiveInfo Size", len(liveInfo)).Tracef("freshLive done")
	return liveInfo, nil
}

func (c *Concern) FindUser(mid int64, load bool) (*UserInfo, error) {
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

func (c *Concern) StatUserWithCache(mid int64, expire time.Duration) (*UserStat, error) {
	userStat, _ := c.StateManager.GetUserStat(mid)
	if userStat != nil {
		return userStat, nil
	}
	resp, err := XRelationStat(mid)
	if err != nil {
		return nil, err
	}
	if resp.GetCode() != 0 {
		return nil, fmt.Errorf("code:%v %v", resp.GetCode(), resp.GetMessage())
	}
	userStat = NewUserStat(mid, resp.GetData().GetFollowing(), resp.GetData().GetFollower())
	err = c.StateManager.AddUserStat(userStat, localdb.ExpireOption(expire))
	if err != nil {
		return nil, err
	}
	return userStat, nil
}

func (c *Concern) ModifyUserRelation(mid int64, act int) (*RelationModifyResponse, error) {
	resp, err := RelationModify(mid, act)
	if err != nil {
		return nil, err
	}
	if resp.GetCode() != 0 {
		logger.WithField("code", resp.GetCode()).
			WithField("message", resp.GetMessage()).
			WithField("act", act).
			WithField("mid", mid).
			Errorf("ModifyUserRelation error")
	} else {
		logger.WithField("mid", mid).WithField("act", act).Debug("modify relation")
	}
	return resp, nil
}

func (c *Concern) SyncSub() {
	defer logger.Debug("SyncSub done")
	resp, err := GetAttentionList()
	if err != nil {
		logger.Errorf("SyncSub error %v", err)
		return
	}
	if resp.GetCode() != 0 {
		logger.WithField("code", resp.GetCode()).
			WithField("msg", resp.GetMessage()).
			Errorf("SyncSub GetAttentionList error")
		return
	}
	var midSet = make(map[int64]bool)
	var attentionMidSet = make(map[int64]bool)
	_, _, _, err = c.List(func(groupCode int64, id interface{}, p concern.Type) bool {
		midSet[id.(int64)] = true
		return true
	})

	if err != nil {
		logger.Errorf("SyncSub List all error %v", err)
		return
	}
	for _, attentionMid := range resp.GetData().GetList() {
		attentionMidSet[attentionMid] = true
	}
	for mid := range midSet {
		if _, found := attentionMidSet[mid]; !found {
			resp, err := c.ModifyUserRelation(mid, ActSub)
			if err == nil && resp.Code == 22002 {
				// 可能是被拉黑了
				logger.WithField("ModifyUserRelation Code", 22002).
					WithField("mid", mid).
					Errorf("ModifyUserRelation failed, remove concern")
				c.RemoveAllById(mid)
			}
			time.Sleep(time.Second * 3)
		}
	}
}

func (c *Concern) FindOrLoadUser(mid int64) (*UserInfo, error) {
	info, err := c.FindUser(mid, false)
	if err != nil || info == nil {
		info, err = c.FindUser(mid, true)
	}
	return info, err
}

func (c *Concern) FindUserLiving(mid int64, load bool) (*LiveInfo, error) {
	userInfo, err := c.FindUser(mid, load)
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

func (c *Concern) FindUserNews(mid int64, load bool) (*NewsInfo, error) {
	userInfo, err := c.FindUser(mid, load)
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

func (c *Concern) GroupWatchNotify(groupCode, mid int64) {
	liveInfo, _ := c.GetLiveInfo(mid)
	if liveInfo.Living() {
		liveInfo.LiveStatusChanged = true
		c.notify <- NewConcernLiveNotify(groupCode, liveInfo)
	}
}
