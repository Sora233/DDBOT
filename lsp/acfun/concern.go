package acfun

import (
	"context"
	"fmt"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/tidwall/buntdb"
	"strconv"
	"strings"
	"time"
)

var logger = utils.GetModuleLogger("acfun-concern")

const (
	Live concern_type.Type = "live"
)

type Concern struct {
	*StateManager
	notify chan<- concern.Notify
}

func (c *Concern) Site() string {
	return Site
}

func (c *Concern) Types() []concern_type.Type {
	return []concern_type.Type{Live}
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (c *Concern) Start() error {
	c.UseNotifyGeneratorFunc(c.notifyGenerator())
	c.UseFreshFunc(c.fresh())
	return c.StateManager.Start()
}

func (c *Concern) Stop() {
	logger.Trace("正在停止acfun concern")
	logger.Trace("正在停止acfun StateManager")
	c.StateManager.Stop()
	logger.Trace("acfun StateManager已停止")
	logger.Trace("acfun concern已停止")
}

func (c *Concern) notifyGenerator() concern.NotifyGeneratorFunc {
	return func(target mt.Target, ievent concern.Event) (result []concern.Notify) {
		log := ievent.Logger()
		switch event := ievent.(type) {
		case *LiveInfo:
			notify := NewConcernLiveNotify(target, event)
			result = append(result, notify)
			if event.Living() {
				log.WithFields(localutils.TargetFields(target)).Trace("living notify")
			} else {
				log.WithFields(localutils.TargetFields(target)).Trace("noliving notify")
			}
		default:
			log.Errorf("unknown concern_type %v", ievent.Type().String())
		}
		return
	}
}

func (c *Concern) fresh() concern.FreshFunc {
	return func(ctx context.Context, eventChan chan<- concern.Event) {
		t := time.NewTimer(time.Second * 3)
		var interval time.Duration
		if config.GlobalConfig != nil {
			interval = config.GlobalConfig.GetDuration("acfun.interval")
		}
		if interval == 0 {
			interval = time.Second * 20
		}
		for {
			select {
			case <-t.C:
			case <-ctx.Done():
				return
			}
			var start = time.Now()
			err := func() error {
				defer func() { logger.WithField("cost", time.Now().Sub(start)).Tracef("watchCore live fresh done") }()

				_, ids, types, err := c.StateManager.ListConcernState(func(target mt.Target, id interface{}, p concern_type.Type) bool {
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
				if len(ids) == 0 {
					// 没有订阅的话，就不要刷新了
					logger.Trace("no concern, skip fresh")
					return nil
				}

				liveInfo, err := c.freshLiveInfo()
				if err != nil {
					return err
				}
				var liveInfoMap = make(map[int64]*LiveInfo)
				for _, info := range liveInfo {
					liveInfoMap[info.Uid] = info
				}

				sendLiveInfo := func(info *LiveInfo) {
					addLiveInfoErr := c.AddLiveInfo(info)
					if addLiveInfoErr != nil {
						// 如果因为系统原因add失败，会造成重复推送
						// 按照ddbot的原则，选择不推送，而非重复推送
						logger.WithField("uid", info.Uid).Errorf("add live info error %v", err)
					} else {
						eventChan <- info
					}
				}
				for _, id := range ids {
					uid := id.(int64)
					oldInfo, _ := c.GetLiveInfo(uid)
					if oldInfo == nil {
						// first live info
						if newInfo, found := liveInfoMap[uid]; found {
							newInfo.liveStatusChanged = true
							sendLiveInfo(newInfo)
						}
						continue
					}
					if !oldInfo.Living() {
						if newInfo, found := liveInfoMap[uid]; found {
							// notliving -> living
							newInfo.liveStatusChanged = true
							sendLiveInfo(newInfo)
						}
					} else {
						if newInfo, found := liveInfoMap[uid]; !found {
							// living -> notliving
							if count := c.IncNotLiveCount(uid); count < 3 {
								logger.WithField("uid", uid).WithField("name", oldInfo.UserInfo.Name).
									WithField("notlive_count", count).
									Debug("notlive counting")
								continue
							} else {
								logger.WithField("uid", uid).WithField("name", oldInfo.UserInfo.Name).
									Debug("notlive count done, notlive confirmed")
							}
							c.ClearNotLiveCount(uid)
							newInfo = &LiveInfo{
								UserInfo:          oldInfo.UserInfo,
								LiveId:            oldInfo.LiveId,
								Title:             oldInfo.Title,
								Cover:             oldInfo.Cover,
								StartTs:           oldInfo.StartTs,
								IsLiving:          false,
								liveStatusChanged: true,
							}
							sendLiveInfo(newInfo)
						} else {
							c.ClearNotLiveCount(uid)
							if newInfo.Title != oldInfo.Title {
								// live title change
								newInfo.liveTitleChanged = true
								sendLiveInfo(newInfo)
							}
						}
					}
				}
				return nil
			}()

			end := time.Now()
			if err == nil {
				logger.WithField("cost", end.Sub(start)).Tracef("watchCore loop done")
			} else {
				logger.WithField("cost", end.Sub(start)).Errorf("watchCore error %v", err)
			}
			t.Reset(interval)
		}
	}
}

func (c *Concern) Add(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	var err error
	var uid = id.(int64)
	log := logger.WithFields(localutils.TargetFields(target)).WithField("id", id)

	err = c.StateManager.CheckTargetConcern(target, id, ctype)
	if err != nil {
		return nil, err
	}
	liveInfo, _ := c.GetLiveInfo(uid)

	userInfo, err := c.FindOrLoadUserInfo(uid)
	if err != nil {
		log.Errorf("FindOrLoadUserInfo error %v", err)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v", id, err)
	}
	_, err = c.StateManager.AddTargetConcern(target, id, ctype)
	if err != nil {
		return nil, err
	}
	err = c.StateManager.SetUidFirstTimestampIfNotExist(uid, time.Now().Add(-time.Second*30).Unix())
	if err != nil && !localdb.IsRollback(err) {
		log.Errorf("SetUidFirstTimestampIfNotExist failed %v", err)
	}
	if ctype.ContainAny(Live) {
		// 其他群关注了同一uid，并且推送过Living，那么给新watch的群也推一份
		if liveInfo != nil && liveInfo.Living() {
			if ctx.GetSource().IsGroup() || ctx.GetSource().IsGuild() {
				defer c.TargetWatchNotify(target, uid)
			}
			if ctx.GetSource().IsPrivate() {
				defer ctx.Send(mmsg.NewText("检测到该用户正在直播，但由于您目前处于私聊模式，" +
					"因此不会在群内推送本次直播，将在该用户下次直播时推送"))
			}
		}
	}
	return concern.NewIdentity(userInfo.Uid, userInfo.GetName()), nil
}

func (c *Concern) Remove(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	mid := id.(int64)
	var identityInfo concern.IdentityInfo
	var allCtype concern_type.Type
	err := c.StateManager.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		identityInfo, _ = c.Get(mid)
		_, err = c.StateManager.RemoveTargetConcern(target, mid, ctype)
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
	if identityInfo == nil {
		identityInfo = concern.NewIdentity(id, "unknown")
	}
	return identityInfo, err
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	return c.FindUserInfo(id.(int64), false)
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func (c *Concern) FindUserInfo(uid int64, load bool) (*UserInfo, error) {
	if load {
		resp, err := LivePage(uid)
		if err != nil {
			return nil, err
		}
		userInfo := &UserInfo{
			Uid:      uid,
			Name:     resp.GetLiveInfo().GetUser().GetName(),
			Followed: int(resp.GetLiveInfo().GetUser().GetFanCountValue()),
			UserImg:  resp.GetLiveInfo().GetUser().GetHeadUrl(),
			LiveUrl:  LiveUrl(uid),
		}
		err = c.AddUserInfo(userInfo)
		if err != nil {
			return nil, err
		}
	}
	return c.StateManager.GetUserInfo(uid)
}

func (c *Concern) FindOrLoadUserInfo(uid int64) (*UserInfo, error) {
	userInfo, _ := c.FindUserInfo(uid, false)
	if userInfo == nil {
		return c.FindUserInfo(uid, true)
	}
	return userInfo, nil
}

func (c *Concern) TargetWatchNotify(target mt.Target, mid int64) {
	liveInfo, _ := c.GetLiveInfo(mid)
	if liveInfo.Living() {
		liveInfo.liveStatusChanged = true
		c.notify <- NewConcernLiveNotify(target, liveInfo)
	}
}

func (c *Concern) freshLiveInfo() ([]*LiveInfo, error) {
	var liveInfos []*LiveInfo
	var pcursor string
	var count = 0
	for pcursor != "no_more" && count < 10 {
		count++
		resp, err := ApiChannelList(100, pcursor)
		if err != nil {
			logger.Errorf("freshLiveInfo error %v", err)
			return nil, err
		}
		pcursor = resp.GetChannelListData().GetPcursor()
		for _, liveItem := range resp.GetChannelListData().GetLiveList() {
			_uid, err := c.ParseId(liveItem.GetUser().GetId())
			if err != nil {
				logger.Errorf("parse id <%v> error %v", liveItem.GetUser().GetId(), err)
				continue
			}
			var cover string
			if len(liveItem.GetCoverUrls()) > 0 {
				cover = liveItem.GetCoverUrls()[0]
			}
			if len(cover) == 0 {
				cover = liveItem.GetUser().GetHeadUrl()
				if pos := strings.Index(cover, "?"); pos > 0 {
					cover = cover[:pos]
				}
			}
			uid := _uid.(int64)
			liveInfos = append(liveInfos, &LiveInfo{
				UserInfo: UserInfo{
					Uid:      uid,
					Name:     liveItem.GetUser().GetName(),
					Followed: int(liveItem.GetUser().GetFanCountValue()),
					UserImg:  liveItem.GetUser().GetHeadUrl(),
					LiveUrl:  LiveUrl(uid),
				},
				LiveId:   liveItem.GetLiveId(),
				Cover:    cover,
				Title:    liveItem.GetTitle(),
				StartTs:  liveItem.GetCreateTime(),
				IsLiving: true,
			})
		}
	}
	if count >= 10 {
		logger.Errorf("ACFUN刷新直播状态分页溢出，是真的有这么多直播吗？如果是真的有这么多直播，可能acfun已经橄榄blive了")
	}
	return liveInfos, nil
}

func NewConcern(notifyChan chan<- concern.Notify) *Concern {
	return &Concern{
		StateManager: NewStateManager(notifyChan),
		notify:       notifyChan,
	}
}
