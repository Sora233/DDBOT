package weibo

import (
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"strconv"
)

var logger = utils.GetModuleLogger("weibo-concern")

type Concern struct {
	*StateManager
}

func (c *Concern) Start() error {
	c.StateManager.UseFreshFunc(c.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) {
		// TODO
		return nil, nil
	}))
	c.StateManager.UseNotifyGeneratorFunc(func(groupCode int64, ievent concern.Event) []concern.Notify {
		var result []concern.Notify
		switch news := ievent.(type) {
		case *NewsInfo:
			for _, n := range NewConcernNewsNotify(groupCode, news) {
				result = append(result, n)
			}
		}
		return result
	})
	return c.StateManager.Start()
}

func (c *Concern) Stop() {
	logger.Tracef("正在停止%v concern", Site)
	logger.Tracef("正在停止%v StateManager", Site)
	c.StateManager.Stop()
	logger.Tracef("%v StateManager已停止", Site)
	logger.Tracef("%v concern已停止", Site)
}

func (c *Concern) Site() string {
	return Site
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (c *Concern) Add(ctx mmsg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(int64)
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("id", id)

	err := c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	info, err := c.FindOrLoadUserInfo(id)
	if err != nil {
		log.Errorf("FindOrLoadUserInfo error %v", err)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v", id, err)
	}
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (c *Concern) Remove(ctx mmsg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(int64)
	identity, _ := c.Get(id)
	_, err := c.StateManager.RemoveGroupConcern(groupCode, id, ctype)
	if identity == nil {
		identity = concern.NewIdentity(_id, "unknown")
	}
	return identity, err
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
	var result = make([]concern.IdentityInfo, 0, len(ids))
	var resultTypes = make([]concern_type.Type, 0, len(ids))
	for index, id := range ids {
		info, err := c.FindOrLoadUserInfo(id.(int64))
		if err != nil {
			log.WithField("id", id.(string)).Errorf("FindOrLoadUserInfo failed %v", err)
			continue
		}
		result = append(result, info)
		resultTypes = append(resultTypes, ctypes[index])
	}
	return result, resultTypes, nil
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	return c.GetUserInfo(id.(int64))
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func (c *Concern) FindUserInfo(uid int64, load bool) (*UserInfo, error) {
	if load {
		profileResp, err := ApiContainerGetIndexProfile(uid)
		if err != nil {
			logger.WithField("uid", uid).Errorf("ApiContainerGetIndexProfile error %v", err)
			return nil, err
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
	c.UseEmitQueue()
	return c
}
