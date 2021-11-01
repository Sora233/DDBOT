package acfun

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("acfun-concern")

const (
	Live concern_type.Type = "live"
)

type concernEvent interface {
	Type() concern_type.Type
}

type Concern struct {
	*StateManager

	wg        sync.WaitGroup
	stop      chan interface{}
	eventChan chan concernEvent
	notify    chan<- concern.Notify
}

func (c *Concern) Site() string {
	return Site
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
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

	go c.watchCore()
	return nil
}

func (c *Concern) notifyLoop() {
	c.wg.Add(1)
	defer c.wg.Done()
	for ievent := range c.eventChan {
		switch ievent.Type() {
		case Live:
			event := ievent.(*LiveInfo)
			log := event.Logger()
			log.Debugf("new event - live notify")

			groups, _, _, err := c.StateManager.ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
				return id.(int64) == event.Uid && p.ContainAny(Live)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}

			for _, groupCode := range groups {
				notify := NewConcernLiveNotify(groupCode, event)
				c.notify <- notify
				if event.Living {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debug("living notify")
				} else {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debug("noliving notify")
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
			break
		}
		var start = time.Now()
		err := func() error {
			defer func() { logger.WithField("cost", time.Now().Sub(start)).Tracef("watchCore live fresh done") }()
			liveInfo, err := c.freshLiveInfo()
			if err != nil {
				return err
			}
			var liveInfoMap = make(map[int64]*LiveInfo)
			for _, info := range liveInfo {
				liveInfoMap[info.Uid] = info
			}

			_, ids, types, err := c.StateManager.ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
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
					logger.WithField("uid", info.Uid).Errorf("add live info error %v", err)
				} else {
					c.eventChan <- info
				}
			}
			for _, id := range ids {
				uid := id.(int64)
				oldInfo, _ := c.GetLiveInfo(uid)
				if oldInfo == nil {
					// first live info
					if newInfo, found := liveInfoMap[uid]; found {
						newInfo.LiveStatusChanged = true
						sendLiveInfo(newInfo)
					}
					continue
				}
				if !oldInfo.Living {
					if newInfo, found := liveInfoMap[uid]; found {
						// notliving -> living
						newInfo.LiveStatusChanged = true
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
							Living:            false,
							LiveStatusChanged: true,
						}
						sendLiveInfo(newInfo)
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

		t.Reset(config.GlobalConfig.GetDuration("acfun.interval"))
	}
}

func (c *Concern) Add(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	var err error
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("id", id)

	err = c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		if err == concern.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}

	liveInfo, err := c.FindOrLoadUserInfo(id.(int64))
	if err != nil {
		log.Errorf("FindOrLoadUserInfo error %v", err)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v", id, err)
	}
	err = c.StateManager.SetUidFirstTimestampIfNotExist(id.(int64), time.Now().Add(-time.Second*30).Unix())
	if err != nil && !localdb.IsRollback(err) {
		log.Errorf("SetUidFirstTimestampIfNotExist failed %v", err)
	}
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(liveInfo.Uid, liveInfo.GetName()), nil
}

func (c *Concern) Remove(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	var info concern.IdentityInfo
	info, _ = c.Get(id)
	_, err := c.StateManager.RemoveGroupConcern(groupCode, id, ctype)
	if info == nil {
		info = concern.NewIdentity(id, "unknown")
	}
	return info, err
}

func (c *Concern) List(groupCode int64, ctype concern_type.Type) ([]concern.IdentityInfo, []concern_type.Type, error) {
	log := logger.WithFields(localutils.GroupLogFields(groupCode))

	_, ids, ctypes, err := c.StateManager.ListConcernState(
		func(_groupCode int64, id interface{}, p concern_type.Type) bool {
			return groupCode == _groupCode && p.ContainAny(ctype)
		})
	if err != nil {
		return nil, nil, err
	}
	var resultTypes = make([]concern_type.Type, 0, len(ids))
	var result = make([]concern.IdentityInfo, 0, len(ids))
	for index, id := range ids {
		userInfo, err := c.FindUserInfo(id.(int64), false)
		if err != nil {
			log.WithField("id", id).Errorf("get FindUserInfo err %v", err)
			continue
		}
		result = append(result, userInfo)
		resultTypes = append(resultTypes, ctypes[index])
	}

	return result, resultTypes, nil
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	return c.FindUserInfo(id.(int64), false)
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func (c *Concern) FindUserInfo(uid int64, load bool) (*UserInfo, error) {
	if load {
		resp, err := V2UserContentProfile(uid)
		if err != nil {
			return nil, err
		}
		if resp.GetErrorid() != 0 {
			return nil, fmt.Errorf("code:%v", resp.GetErrorid())
		}
		userInfo := &UserInfo{
			Uid:      resp.GetVdata().GetUserId(),
			Name:     resp.GetVdata().GetUsername(),
			Followed: int(resp.GetVdata().GetFollowed()),
			UserImg:  resp.GetVdata().GetUserImg(),
			LiveUrl:  LiveUrl(resp.GetVdata().GetUserId()),
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

func (c *Concern) freshLiveInfo() ([]*LiveInfo, error) {
	var liveInfos []*LiveInfo
	var pcursor string
	for pcursor != "no_more" {
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
				if pos := strings.Index(cover, "?"); pos != -1 {
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
				LiveId:  liveItem.GetLiveId(),
				Cover:   cover,
				Title:   liveItem.GetTitle(),
				StartTs: liveItem.GetCreateTime(),
				Living:  true,
			})
		}
	}
	return liveInfos, nil
}

func (c *Concern) FreshIndex(groupCode ...int64) {
	c.StateManager.FreshIndex(groupCode...)
	localdb.CreatePatternIndex(c.UserInfoKey, nil)
	localdb.CreatePatternIndex(c.LiveInfoKey, nil)
}

func NewConcern(notifyChan chan<- concern.Notify) *Concern {
	return &Concern{
		StateManager: NewStateManager(),
		eventChan:    make(chan concernEvent, 16),
		notify:       notifyChan,
	}
}
