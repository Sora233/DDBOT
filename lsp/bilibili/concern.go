package bilibili

import (
	"context"
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("bilibili-concern")

const (
	Live concern_type.Type = "live"
	News concern_type.Type = "news"
)

type Concern struct {
	*StateManager
	unsafeStart atomic.Bool
	notify      chan<- concern.Notify
	stop        chan interface{}
	wg          sync.WaitGroup
}

func (c *Concern) Site() string {
	return "bilibili"
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return ParseUid(s)
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		notify: notify,
		stop:   make(chan interface{}),
	}
	c.StateManager = NewStateManager(c)
	lastFresh, _ := c.GetLastFreshTime()
	if lastFresh > 0 && time.Now().Sub(time.Unix(lastFresh, 0)) > time.Minute*30 {
		c.unsafeStart.Store(true)
		time.AfterFunc(time.Minute*3, func() {
			c.unsafeStart.Store(false)
		})
	}
	return c
}

func (c *Concern) Stop() {
	logger.Trace("正在停止bilibili concern")
	if c.stop != nil {
		close(c.stop)
	}
	logger.Trace("正在停止bilibili StateManager")
	c.StateManager.Stop()
	logger.Trace("bilibili StateManager已停止")
	c.wg.Wait()
	logger.Trace("bilibili concern已停止")
}

func (c *Concern) Start() error {
	Init()

	if !IsVerifyGiven() {
		logger.Errorf("注意：B站配置不完整，B站相关功能无法使用！")
		return nil
	}

	c.StateManager.UseNotifyGeneratorFunc(c.notifyGenerator())
	c.StateManager.UseFreshFunc(c.fresh())

	go func() {
		c.wg.Add(1)
		defer c.wg.Done()
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
	return c.StateManager.Start()
}

func (c *Concern) Add(ctx mmsg.IMsgCtx,
	groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	mid := _id.(int64)
	selfUid := accountUid.Load()
	var err error
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("mid", mid)

	err = c.StateManager.CheckGroupConcern(groupCode, mid, ctype)
	if err != nil {
		return nil, err
	}
	var userInfo *UserInfo
	var liveInfo *LiveInfo

	liveInfo, _ = c.GetLiveInfo(mid)

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
		if userStat.Follower == 0 {
			return nil, fmt.Errorf("该用户粉丝数为0，请确认您的订阅目标是否正确，注意使用UID而非直播间ID")
		}
		userInfo.UserStat = userStat
	}

	if selfUid != 0 && mid != selfUid {
		oldCtype, err := c.StateManager.GetConcern(mid)
		if err != nil {
			log.Errorf("GetConcern error %v", err)
		} else if oldCtype.Empty() {
			var actType = ActSub
			if config.GlobalConfig.GetBool("bilibili.hiddenSub") {
				actType = ActHiddenSub
			}
			resp, err := c.ModifyUserRelation(mid, actType)
			if err != nil {
				if err == ErrVerifyRequired {
					log.Errorf("ModifyUserRelation error %v", err)
					return nil, fmt.Errorf("关注用户失败 - 未配置B站")
				} else {
					log.WithField("action", actType).Errorf("ModifyUserRelation error %v", err)
					return nil, fmt.Errorf("关注用户失败 - 内部错误")
				}
			}
			if resp.Code != 0 {
				return nil, fmt.Errorf("关注用户失败 - %v", resp.GetMessage())
			}
		}
	} else {
		log.Debug("正在订阅账号自己，跳过关注")
	}

	_, err = c.StateManager.AddGroupConcern(groupCode, mid, ctype)
	if err != nil {
		log.Errorf("AddGroupConcern error %v", err)
		return nil, fmt.Errorf("关注用户失败 - 内部错误")
	}
	err = c.StateManager.SetUidFirstTimestampIfNotExist(mid, time.Now().Add(-time.Second*30).Unix())
	if err != nil && !localdb.IsRollback(err) {
		log.Errorf("SetUidFirstTimestampIfNotExist failed %v", err)
	}
	_ = c.StateManager.AddUserInfo(userInfo)
	if ctype.ContainAny(Live) {
		// 其他群关注了同一uid，并且推送过Living，那么给新watch的群也推一份
		if liveInfo != nil && liveInfo.Living() {
			if ctx.GetTarget().TargetType().IsGroup() {
				defer c.GroupWatchNotify(groupCode, mid)
			}
			if ctx.GetTarget().TargetType().IsPrivate() {
				defer ctx.Send(mmsg.NewText("检测到该用户正在直播，但由于您目前处于私聊模式，" +
					"因此不会在群内推送本次直播，将在该用户下次直播时推送"))
			}
		}
	}
	const followerCap = 50
	if userInfo != nil &&
		userInfo.UserStat != nil &&
		ctype.ContainAny(Live) &&
		userInfo.UserStat.Follower < followerCap {
		ctx.Send(mmsg.NewTextf("注意：检测到用户【%v】粉丝数少于%v，"+
			"请确认您的订阅目标是否正确，注意使用UID而非直播间ID", userInfo.Name, followerCap))
	}

	return concern.NewIdentity(mid, userInfo.GetName()), nil
}

func (c *Concern) Remove(ctx mmsg.IMsgCtx,
	groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	mid := id.(int64)
	var identityInfo concern.IdentityInfo
	var allCtype concern_type.Type
	err := c.StateManager.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		identityInfo, _ = c.Get(mid)
		_, err = c.StateManager.RemoveGroupConcern(groupCode, mid, ctype)
		if err != nil {
			return err
		}
		allCtype, err = c.StateManager.GetConcern(mid)
		if err != nil {
			return err
		}
		// 如果已经没有watch live的了，此时应该把liveinfo删掉，否则会无法刷新到livelinfo
		// 如果此时liveinfo是living状态，则此状态会一直保留，下次watch时会以为在living错误推送
		if !allCtype.ContainAll(Live) {
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
	if err == nil && config.GlobalConfig != nil {
		if config.GlobalConfig.GetBool("bilibili.unsub") && allCtype.Empty() {
			c.unsubUser(mid)
		}
	}
	if identityInfo == nil {
		identityInfo = concern.NewIdentity(id, "unknown")
	}
	return identityInfo, err
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	userInfo, err := c.FindUser(id.(int64), false)
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(id, userInfo.GetName()), nil
}

func (c *Concern) List(groupCode int64, ctype concern_type.Type) ([]concern.IdentityInfo, []concern_type.Type, error) {
	log := logger.WithFields(localutils.GroupLogFields(groupCode))

	_, mids, ctypes, err := c.StateManager.ListConcernState(
		func(_groupCode int64, id interface{}, p concern_type.Type) bool {
			return groupCode == _groupCode && p.ContainAny(ctype)
		})
	if err != nil {
		return nil, nil, err
	}
	mids, ctypes, err = c.StateManager.GroupTypeById(mids, ctypes)
	if err != nil {
		return nil, nil, err
	}
	var result = make([]concern.IdentityInfo, 0, len(mids))
	var resultTypes = make([]concern_type.Type, 0, len(mids))
	for index, mid := range mids {
		userInfo, err := c.StateManager.GetUserInfo(mid.(int64))
		if err != nil {
			log.WithField("mid", mid).Errorf("GetUserInfo error %v", err)
			continue
		}
		result = append(result, concern.NewIdentity(userInfo.Mid, userInfo.GetName()))
		resultTypes = append(resultTypes, ctypes[index])
	}
	return result, resultTypes, nil
}

func (c *Concern) notifyGenerator() concern.NotifyGeneratorFunc {
	return func(groupCode int64, ievent concern.Event) (result []concern.Notify) {
		log := ievent.Logger()
		switch event := ievent.(type) {
		case *LiveInfo:
			if event.Status == LiveStatus_Living {
				log.WithFields(localutils.GroupLogFields(groupCode)).Debug("living notify")
			} else if event.Status == LiveStatus_NoLiving {
				log.WithFields(localutils.GroupLogFields(groupCode)).Debug("noliving notify")
			} else {
				log.WithFields(localutils.GroupLogFields(groupCode)).Error("unknown live status")
			}
			result = append(result, NewConcernLiveNotify(groupCode, event))
		case *NewsInfo:
			notifies := NewConcernNewsNotify(groupCode, event, c)
			log.WithFields(localutils.GroupLogFields(groupCode)).
				WithField("Size", len(notifies)).Debug("news notify")
			for _, notify := range notifies {
				result = append(result, notify)
			}
		}
		return
	}
}

// fresh 这个fresh不能启动多个
func (c *Concern) fresh() concern.FreshFunc {
	return func(ctx context.Context, eventChan chan<- concern.Event) {
		t := time.NewTimer(time.Second * 3)
		var interval time.Duration
		if config.GlobalConfig == nil {
			interval = time.Second * 20
		} else {
			interval = config.GlobalConfig.GetDuration("bilibili.interval")
		}
		for {
			select {
			case <-t.C:
			case <-ctx.Done():
				return
			}
			start := time.Now()
			var errGroup errgroup.Group

			errGroup.Go(func() error {
				defer func() {
					logger.WithField("cost", time.Now().Sub(start)).
						Tracef("watchCore dynamic fresh done")
				}()
				newsList, err := c.freshDynamicNew()
				if err != nil {
					logger.Errorf("freshDynamicNew failed %v", err)
					return err
				} else {
					for _, news := range newsList {
						eventChan <- news
					}
				}
				return nil
			})

			errGroup.Go(func() error {
				defer func() {
					logger.WithField("cost", time.Now().Sub(start)).
						Tracef("watchCore live fresh done")
				}()
				liveInfo, err := c.freshLive()
				if err != nil {
					logger.Errorf("freshLive error %v", err)
					return err
				}
				// liveInfoMap内是所有正在直播的列表，没有直播的不应该放进去
				var liveInfoMap = make(map[int64]*LiveInfo)
				for _, info := range liveInfo {
					liveInfoMap[info.Mid] = info
				}

				_, ids, types, err := c.StateManager.ListConcernState(
					func(groupCode int64, id interface{}, p concern_type.Type) bool {
						return p.ContainAny(Live)
					})
				if err != nil {
					logger.Errorf("ListConcernState error %v", err)
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
						eventChan <- info
					}
				}

				selfUid := accountUid.Load()
				for _, id := range ids {
					mid := id.(int64)
					if selfUid != 0 && selfUid == mid {
						// 特殊处理下关注自己
						accResp, err := XSpaceAccInfo(selfUid)
						if err != nil {
							logger.Errorf("freshLive self-fresh %v error %v", selfUid, err)
							return err
						}
						liveRoom := accResp.GetData().GetLiveRoom()
						selfLiveInfo := NewLiveInfo(
							NewUserInfo(selfUid, liveRoom.GetRoomid(), accResp.GetData().GetName(), liveRoom.GetUrl()),
							liveRoom.GetTitle(),
							liveRoom.GetCover(),
							liveRoom.GetLiveStatus(),
						)
						if selfLiveInfo.Living() {
							liveInfoMap[selfUid] = selfLiveInfo
						}
					}
					oldInfo, _ := c.GetLiveInfo(mid)
					if oldInfo == nil {
						// first live info
						if newInfo, found := liveInfoMap[mid]; found {
							newInfo.liveStatusChanged = true
							sendLiveInfo(newInfo)
						}
						continue
					}
					if oldInfo.Status == LiveStatus_NoLiving {
						if newInfo, found := liveInfoMap[mid]; found {
							// notliving -> living
							newInfo.liveStatusChanged = true
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
							newInfo = NewLiveInfo(&oldInfo.UserInfo, oldInfo.LiveTitle,
								oldInfo.Cover, LiveStatus_NoLiving)
							newInfo.liveStatusChanged = true
							sendLiveInfo(newInfo)
						} else {
							if newInfo.LiveTitle == "bilibili主播的直播间" {
								newInfo.LiveTitle = oldInfo.LiveTitle
							}
							c.ClearNotLiveCount(mid)
							if newInfo.LiveTitle != oldInfo.LiveTitle {
								// live title change
								newInfo.liveTitleChanged = true
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
				c.SetLastFreshTime(time.Now().Unix())
			} else {
				logger.WithField("cost", end.Sub(start)).Errorf("watchCore error %v", err)
			}
			t.Reset(interval)
		}
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
	logger.WithField("cost", time.Now().Sub(start)).
		WithField("NewsInfo Size", len(result)).
		Trace("freshDynamicNew done")
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
			if resp.GetCode() == -101 && strings.Contains(resp.GetMessage(), "未登录") {
				logger.Errorf("刷新直播列表失败，可能是cookie失效，将尝试重新获取cookie")
				ClearCookieInfo(username)
				atomicVerifyInfo.Store(new(VerifyInfo))
			} else if resp.GetCode() == -400 {
				logger.Errorf("刷新直播列表失败，可能是自动登陆失败，请查看文档尝试手动设置b站cookie")
			} else {
				logger.Errorf("freshLive FeedList code %v msg %v", resp.GetCode(), resp.GetMessage())
			}
			return nil, fmt.Errorf("freshLive FeedList error code %v msg %v", resp.GetCode(), resp.GetMessage())
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
		if int32(page) > maxPage {
			break
		}
		if zeroCount >= 3 {
			// 认为是真的无人在直播，可能是关注比较少
			if maxPage > 1 {
				logger.WithFields(logrus.Fields{
					"Page":          page,
					"MaxPage":       maxPage,
					"LiveInfo Size": len(liveInfo),
				}).Errorf("直播信息刷新异常结束，如果该信息没有频繁出现，可以忽略。")
			}
			break
		}
	}
	logger.WithFields(logrus.Fields{
		"cost":          time.Since(start),
		"Page":          page,
		"MaxPage":       maxPage,
		"LiveInfo Size": len(liveInfo),
	}).Tracef("freshLive done")
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
		newLiveInfo := NewLiveInfo(newUserInfo,
			resp.GetData().GetLiveRoom().GetTitle(),
			resp.GetData().GetLiveRoom().GetCover(),
			resp.GetData().GetLiveRoom().GetLiveStatus(),
		)
		// AddLiveInfo 会顺便添加UserInfo
		err = c.StateManager.AddLiveInfo(newLiveInfo)
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
	err = c.StateManager.AddUserStat(userStat, expire)
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
	_, _, _, err = c.StateManager.ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
		midSet[id.(int64)] = true
		return true
	})

	if err != nil {
		logger.Errorf("SyncSub ListConcernState all error %v", err)
		return
	}
	for _, attentionMid := range resp.GetData().GetList() {
		attentionMidSet[attentionMid] = true
	}

	var actType = ActSub
	if config.GlobalConfig.GetBool("bilibili.hiddenSub") {
		actType = ActHiddenSub
	}

	for mid := range midSet {
		if mid == accountUid.Load() {
			continue
		}
		if _, found := attentionMidSet[mid]; !found {
			resp, err := c.ModifyUserRelation(mid, actType)
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
	_, err := c.FindUser(mid, load)
	if err != nil {
		return nil, err
	}
	// FindUser会顺便刷新LiveInfo，所以这里不用再刷新了
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
		liveInfo.liveStatusChanged = true
		c.notify <- NewConcernLiveNotify(groupCode, liveInfo)
	}
}

func (c *Concern) RemoveAllByGroupCode(groupCode int64) ([]string, error) {
	keys, err := c.StateManager.RemoveAllByGroupCode(groupCode)
	if config.GlobalConfig != nil && config.GlobalConfig.GetBool("bilibili.unsub") {
		var changedIdSet = make(map[int64]interface{})
		if err == nil {
			for _, key := range keys {
				if !strings.HasPrefix(key, c.GroupConcernStateKey()) {
					continue
				}
				_, id, err := c.ParseGroupConcernStateKey(key)
				if err != nil {
					continue
				}
				changedIdSet[id.(int64)] = true
			}
			c.RWCover(func() error {
				for mid := range changedIdSet {
					ctype, err := c.GetConcern(mid)
					if err != nil {
						continue
					}
					if !ctype.ContainAll(Live) {
						c.StateManager.DeleteLiveInfo(mid)
					}
				}
				return nil
			})
		}
		go func() {
			// 考虑到unsub是个网络操作，还是不要占用事务了
			for mid := range changedIdSet {
				if ctype, err := c.GetConcern(mid); err == nil && ctype.Empty() {
					c.unsubUser(mid)
				}
			}
		}()
	}
	return keys, err
}

func (c *Concern) unsubUser(mid int64) {
	resp, err := c.ModifyUserRelation(mid, ActUnsub)
	if err != nil {
		logger.Errorf("取消关注失败 - %v", err)
	} else if resp.GetCode() != 0 {
		logger.Errorf("取消关注失败 - %v - %v", resp.GetCode(), resp.GetMessage())
	} else {
		logger.WithField("mid", mid).Info("取消关注成功")
	}
}
