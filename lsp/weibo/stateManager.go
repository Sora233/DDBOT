package weibo

import (
	"errors"
	"time"

	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	"github.com/Sora233/DDBOT/v2/lsp/concern"
)

type StateManager struct {
	*concern.StateManager
	extraKeySet
}

func NewStateManager(notify chan<- concern.Notify) *StateManager {
	return &StateManager{
		StateManager: concern.NewStateManagerWithInt64ID(Site, notify),
	}
}

func (s *StateManager) AddUserInfo(info *UserInfo) error {
	if info == nil {
		return errors.New("<nil userInfo>")
	}
	return s.SetJson(s.UserInfoKey(info.Uid), info)
}

func (s *StateManager) GetUserInfo(uid int64) (*UserInfo, error) {
	var userInfo *UserInfo
	err := s.GetJson(s.UserInfoKey(uid), &userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (s *StateManager) AddNewsInfo(info *NewsInfo) error {
	if info == nil {
		return errors.New("<nil newsInfo>")
	}
	return s.RWCover(func() error {
		var err error
		err = s.SetJson(s.UserInfoKey(info.Uid), info.UserInfo)
		if err != nil {
			return err
		}
		return s.SetJson(s.NewsInfoKey(info.Uid), info)
	})
}

func (s *StateManager) GetNewsInfo(uid int64) (*NewsInfo, error) {
	var newsInfo *NewsInfo
	err := s.GetJson(s.NewsInfoKey(uid), &newsInfo)
	if err != nil {
		return nil, err
	}
	return newsInfo, nil
}

func (s *StateManager) MarkMblogId(mblogId string) (replaced bool, err error) {
	err = s.Set(s.MarkMblogIdKey(mblogId), "",
		localdb.SetExpireOpt(time.Hour*120), localdb.SetGetIsOverwriteOpt(&replaced))
	return
}
