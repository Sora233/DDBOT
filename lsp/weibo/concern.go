package weibo

import (
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/tidwall/buntdb"
	"strconv"
	"time"
)

var logger = utils.GetModuleLogger("weibo-concern")

type Concern struct {
	*StateManager
}

func (c *Concern) Site() string {
	return Site
}

func (c *Concern) Types() []concern_type.Type {
	return []concern_type.Type{News}
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func (c *Concern) Start() error {
	c.UseEmitQueue()
	c.StateManager.UseFreshFunc(c.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) {
		uid := id.(int64)
		if p.ContainAny(News) {
			newsInfo, err := c.freshNews(uid)
			if err != nil {
				return nil, err
			}
			if len(newsInfo.Cards) == 0 {
				return nil, nil
			}
			return []concern.Event{newsInfo}, nil

		}
		return nil, nil
	}))
	c.StateManager.UseNotifyGeneratorFunc(c.notifyGenerator())
	return c.StateManager.Start()
}

func (c *Concern) Stop() {
	logger.Tracef("正在停止%v concern", Site)
	logger.Tracef("正在停止%v StateManager", Site)
	c.StateManager.Stop()
	logger.Tracef("%v StateManager已停止", Site)
	logger.Tracef("%v concern已停止", Site)
}

func (c *Concern) Add(ctx mmsg.IMsgCtx, target mt.Target, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(int64)
	log := logger.WithFields(localutils.TargetFields(target)).WithField("id", id)

	err := c.StateManager.CheckTargetConcern(target, id, ctype)
	if err != nil {
		return nil, err
	}
	info, err := c.FindOrLoadUserInfo(id)
	if err != nil {
		log.Errorf("FindOrLoadUserInfo error %v", err)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v", id, err)
	}
	if r, _ := c.GetStateManager().GetTargetConcern(target, id); r.Empty() {
		cardResp, err := ApiContainerGetIndexCards(id)
		if err != nil {
			log.Errorf("ApiContainerGetIndexCards error %v", err)
			return nil, fmt.Errorf("添加订阅失败 - 刷新用户微博失败")
		}
		if cardResp.GetOk() != 1 {
			log.WithField("respOk", cardResp.GetOk()).
				WithField("respMsg", cardResp.GetMsg()).
				Errorf("ApiContainerGetIndexCards not ok")
			return nil, fmt.Errorf("添加订阅失败 - 无法查看用户微博")
		}
		// LatestNewsTs 第一次就手动塞一下时间戳，以此来过滤旧的微博
		err = c.AddNewsInfo(&NewsInfo{
			UserInfo:     info,
			LatestNewsTs: time.Now().Unix(),
		})
		if err != nil {
			log.Errorf("AddNewsInfo error %v", err)
			return nil, fmt.Errorf("添加订阅失败 - 内部错误")
		}
	}
	_, err = c.StateManager.AddTargetConcern(target, id, ctype)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (c *Concern) Remove(ctx mmsg.IMsgCtx, target mt.Target, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(int64)
	identity, _ := c.Get(id)
	_, err := c.StateManager.RemoveTargetConcern(target, id, ctype)
	if identity == nil {
		identity = concern.NewIdentity(_id, "unknown")
	}
	return identity, err
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	return c.GetUserInfo(id.(int64))
}

func (c *Concern) freshNews(uid int64) (*NewsInfo, error) {
	log := logger.WithField("uid", uid)
	userInfo, err := c.FindOrLoadUserInfo(uid)
	if err != nil {
		return nil, fmt.Errorf("FindOrLoadUserInfo error %v", err)
	}
	if userInfo == nil {
		return nil, fmt.Errorf("userInfo is nil")
	}
	cardResp, err := ApiContainerGetIndexCards(uid)
	if err != nil {
		log.Errorf("ApiContainerGetIndexCards error %v", err)
		return nil, err
	}
	if cardResp.GetOk() != 1 {
		log.WithField("respOk", cardResp.GetOk()).
			WithField("respMsg", cardResp.GetMsg()).
			Errorf("ApiContainerGetIndexCards not ok")
		return nil, errors.New("ApiContainerGetIndexCards not success")
	}
	var lastTs int64
	var newsInfo = &NewsInfo{UserInfo: userInfo}
	oldNewsInfo, err := c.GetNewsInfo(uid)
	if err == buntdb.ErrNotFound {
		lastTs = time.Now().Unix()
		newsInfo.LatestNewsTs = lastTs
	} else {
		lastTs = oldNewsInfo.LatestNewsTs
		newsInfo.LatestNewsTs = lastTs
	}
	var replaced bool
	for _, card := range cardResp.GetData().GetCards() {
		replaced, err = c.MarkMblogId(card.GetMblog().GetId())
		if err != nil || replaced {
			if err != nil {
				log.WithField("mblogId", card.GetMblog().GetId()).
					Errorf("MarkMblogId error %v", err)
			}
			continue
		}
		if t, err := time.Parse(time.RubyDate, card.GetMblog().GetCreatedAt()); err != nil {
			log.WithField("time_string", card.GetMblog().GetCreatedAt()).
				Errorf("can not parse Mblog.CreatedAt %v", err)
			continue
		} else if lastTs > 0 && t.Unix() > lastTs {
			newsInfo.Cards = append(newsInfo.Cards, card)
			if t.Unix() > newsInfo.LatestNewsTs {
				newsInfo.LatestNewsTs = t.Unix()
			}
		}
	}
	err = c.AddNewsInfo(newsInfo)
	if err != nil {
		log.Errorf("AddNewsInfo error %v", err)
		return nil, err
	}
	return newsInfo, nil
}

func (c *Concern) notifyGenerator() concern.NotifyGeneratorFunc {
	return func(target mt.Target, ievent concern.Event) []concern.Notify {
		var result []concern.Notify
		switch news := ievent.(type) {
		case *NewsInfo:
			if len(news.Cards) > 0 {
				for _, n := range NewConcernNewsNotify(target, news) {
					result = append(result, n)
				}
			}
		}
		return result
	}
}

func (c *Concern) FindUserInfo(uid int64, load bool) (*UserInfo, error) {
	if load {
		profileResp, err := ApiContainerGetIndexProfile(uid)
		if err != nil {
			logger.WithField("uid", uid).Errorf("ApiContainerGetIndexProfile error %v", err)
			return nil, err
		}
		if profileResp.GetOk() != 1 {
			logger.WithField("respOk", profileResp.GetOk()).
				WithField("respMsg", profileResp.GetMsg()).
				Errorf("ApiContainerGetIndexProfile not ok")
			return nil, errors.New("接口请求失败")
		}
		err = c.AddUserInfo(&UserInfo{
			Uid:             uid,
			Name:            profileResp.GetData().GetUserInfo().GetScreenName(),
			ProfileImageUrl: profileResp.GetData().GetUserInfo().GetProfileImageUrl(),
			ProfileUrl:      profileResp.GetData().GetUserInfo().GetProfileUrl(),
		})
		if err != nil {
			logger.WithField("uid", uid).Errorf("AddUserInfo error %v", err)
		}
	}
	return c.GetUserInfo(uid)
}

func (c *Concern) FindOrLoadUserInfo(uid int64) (*UserInfo, error) {
	info, _ := c.FindUserInfo(uid, false)
	if info == nil {
		return c.FindUserInfo(uid, true)
	}
	return info, nil
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		StateManager: NewStateManager(notify),
	}
	return c
}
