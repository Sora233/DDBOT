package bilibili

import (
	"fmt"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/DDBOT/utils/expirable"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/tidwall/buntdb"
	"go.uber.org/atomic"
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
	attentionListExpirable *expirable.Expirable
	unsafeStart            atomic.Bool
	notify                 chan<- concern.Notify
	stop                   chan interface{}
	wg                     sync.WaitGroup
	cacheStartTs           int64
}

func (c *Concern) Site() string {
	return Site
}

func (c *Concern) Types() []concern_type.Type {
	return []concern_type.Type{Live, News}
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return ParseUid(s)
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		notify:       notify,
		stop:         make(chan interface{}),
		cacheStartTs: time.Now().Unix(),
		attentionListExpirable: expirable.NewExpirable(time.Second*20, func() interface{} {
			var m = make(map[int64]interface{})
			resp, err := GetAttentionList()
			if err != nil {
				logger.Errorf("GetAttentionList error %v", err)
				return m
			}
			if resp.GetCode() != 0 {
				logger.Errorf("GetAttentionList error %v - %v", resp.GetCode(), resp.GetMessage())
				return m
			}
			for _, id := range resp.GetData().GetList() {
				m[id] = struct{}{}
			}
			return m
		}),
	}
	c.StateManager = NewStateManager(c)
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
	lastFresh, _ := c.GetLastFreshTime()
	if lastFresh > 0 && time.Now().Sub(time.Unix(lastFresh, 0)) > time.Minute*30 {
		logger.Debug("Unsafe Start Mode")
		c.unsafeStart.Store(true)
		time.AfterFunc(time.Minute*3, func() {
			c.unsafeStart.Store(false)
		})
	}
	c.UseNotifyGeneratorFunc(c.notifyGenerator())
	if !IsVerifyGiven() {
		logger.Warnf("未设置B站账户，将使用慢速模式，推荐订阅数量不超过5个，否则推送将出现较长延迟，如需更多订阅，推荐您配置使用B站账号，最高可支持2000订阅。")
		c.UseEmitQueue()
		c.UseFreshFunc(c.emitQueueFresher())
	} else {
		c.UseFreshFunc(c.fresh())
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
	}
	return c.StateManager.Start()
}

func (c *Concern) Add(ctx mmsg.IMsgCtx,
	groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	mid := _id.(int64)
	selfUid := accountUid.Load()
	var watchSelf = selfUid != 0 && selfUid == mid
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

	if !c.EmitQueueEnabled() {
		if !c.checkRelation(mid) {
			userStat, err := c.StatUserWithCache(mid, time.Second*20)
			if err != nil {
				log.Errorf("get UserStat error %v\n", err)
			} else if userStat != nil {
				var minFollowerCap = cfg.GetBilibiliMinFollowerCap()
				if !watchSelf && userStat.Follower <= int64(minFollowerCap) {
					return nil, fmt.Errorf("订阅目标粉丝数未超过%v无法订阅，请确认您的订阅目标是否正确，注意使用UID而非直播间ID", minFollowerCap)
				}
				userInfo.UserStat = userStat
			}
		}

		if !watchSelf {
			oldCtype, err := c.StateManager.GetConcern(mid)
			if err != nil {
				log.Errorf("GetConcern error %v", err)
			} else if oldCtype.Empty() {
				if c.checkRelation(mid) {
					log.Infof("当前B站账户已关注该用户，跳过关注")
				} else {
					if cfg.GetBilibiliDisableSub() {
						return nil, fmt.Errorf("关注用户失败 - 该用户未在关注列表内，请联系管理员")
					}
					var actType = ActSub
					if cfg.GetBilibiliHiddenSub() {
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
						if resp.Code == 22015 {
							log.Errorf("关注用户失败 %v - %v | 请尝试手动登陆b站账户关注该用户，如果您已手动关注该用户，请在20秒后重试", resp.GetCode(), resp.GetMessage())
							return nil, fmt.Errorf("关注用户失败 - %v", resp.GetMessage())
						}
						log.Errorf("关注用户失败 %v - %v", resp.GetCode(), resp.GetMessage())
						return nil, fmt.Errorf("关注用户失败 - %v", resp.GetMessage())
					}
				}
			}
		} else if selfUid != 0 {
			log.Debug("正在订阅账号自己，跳过关注")
		}
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
	if userInfo != nil &&
		userInfo.UserStat != nil &&
		ctype.ContainAny(Live) &&
		userInfo.UserStat.Follower < followerNotifyCap {
		ctx.Send(mmsg.NewTextf("注意：检测到用户【%v】粉丝数少于%v，"+
			"请确认您的订阅目标是否正确，注意使用UID而非直播间ID", userInfo.Name, followerNotifyCap))
	}

	return userInfo, nil
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
		if cfg.GetBilibiliUnsub() && allCtype.Empty() {
			c.unsubUser(mid)
		}
	}
	if identityInfo == nil {
		identityInfo = concern.NewIdentity(id, "unknown")
	}
	return identityInfo, err
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	return c.FindUser(id.(int64), false)
}

func (c *Concern) notifyGenerator() concern.NotifyGeneratorFunc {
	return func(groupCode int64, ievent concern.Event) (result []concern.Notify) {
		log := ievent.Logger()
		switch event := ievent.(type) {
		case *LiveInfo:
			if event.Status == LiveStatus_Living {
				log.WithFields(localutils.GroupLogFields(groupCode)).Trace("living notify")
			} else if event.Status == LiveStatus_NoLiving {
				log.WithFields(localutils.GroupLogFields(groupCode)).Trace("noliving notify")
			} else {
				log.WithFields(localutils.GroupLogFields(groupCode)).Error("unknown live status")
			}
			result = append(result, NewConcernLiveNotify(groupCode, event))
		case *NewsInfo:
			notifies := NewConcernNewsNotify(groupCode, event, c)
			log.WithFields(localutils.GroupLogFields(groupCode)).
				WithField("Size", len(notifies)).Trace("news notify")
			for _, notify := range notifies {
				result = append(result, notify)
			}
		}
		return
	}
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
	var resp *RelationModifyResponse
	var err error
	// b站好像有新灰度，-111代表 csrf校验失败
	// 只有shjd这个idc会返回这个错误
	// 当返回-111的时候重试一下
	localutils.Retry(3, time.Millisecond*300, func() bool {
		resp, err = RelationModify(mid, act)
		return err != nil || resp.GetCode() != -111
	})
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

	var disableSub = false
	if config.GlobalConfig.GetBool("bilibili.disableSub") {
		disableSub = true
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
			if disableSub {
				logger.Warnf("检测到存在未关注的订阅目标 UID:%v，同时禁用了b站自动关注，将无法推送该用户", mid)
				continue
			}
			resp, err := c.ModifyUserRelation(mid, actType)
			if err == nil {
				switch resp.Code {
				case 22002, 22003, 22013:
					// 22002 可能是被拉黑了
					// 22003 主动拉黑对方
					// 22013 帐号注销
					logger.WithField("ModifyUserRelation Code", resp.Code).
						WithField("ModifyUserRelation Message", resp.Message).
						WithField("mid", mid).
						Errorf("ModifyUserRelation failed, remove concern")
					c.RemoveAllById(mid)
				}
			} else {
				logger.Errorf("ModifyUserRelation error %v", err)
			}
			time.Sleep(time.Second * 3)
			select {
			case <-c.stop:
				return
			default:
			}
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
	userInfo, err := c.FindOrLoadUser(mid)
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
	if cfg.GetBilibiliUnsub() {
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

func (c *Concern) checkRelation(mid int64) bool {
	var atr = c.attentionListExpirable.Do()
	if atr == nil {
		return false
	}
	var matr = atr.(map[int64]interface{})
	if _, found := matr[mid]; found {
		return true
	} else {
		return false
	}
}

func (c *Concern) filterCard(card *Card) bool {
	// 2021-08-15 发现好像是系统推荐的直播间，非人为操作
	// 在event阶段过滤掉
	if card.GetDesc().GetType() == DynamicDescType_WithLiveV2 {
		return false
	}
	uid := card.GetDesc().GetUid()
	// 应该用dynamic_id_str
	// 但好像已经没法保持向后兼容同时改动了
	// 只能相信概率论了，出问题的概率应该比较小，出问题会导致推送丢失
	replaced, err := c.MarkDynamicId(card.GetDesc().GetDynamicId())
	if err != nil {
		logger.WithField("uid", uid).
			WithField("dynamicId", card.GetDesc().GetDynamicId()).
			Errorf("MarkDynamicId error %v", err)
		return false
	}
	if replaced {
		return false
	}
	var tsLimit int64
	if cfg.GetBilibiliOnlyOnlineNotify() {
		tsLimit = c.cacheStartTs
	} else {
		tsLimit, err = c.StateManager.GetUidFirstTimestamp(uid)
		if err != nil {
			return true
		}
	}
	if card.GetDesc().GetTimestamp() < tsLimit {
		logger.WithField("uid", uid).
			WithField("dynamicId", card.GetDesc().GetDynamicId()).
			Trace("past news skip")
		return false
	}
	return true
}
