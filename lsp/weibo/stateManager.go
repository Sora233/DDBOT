package weibo

import (
	"errors"
	"github.com/Sora233/DDBOT/lsp/concern"
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
	return s.JsonSave(s.UserInfoKey(info.Uid), info, true)
}

func (s *StateManager) GetUserInfo(uid int64) (*UserInfo, error) {
	var userInfo *UserInfo
	err := s.JsonGet(s.UserInfoKey(uid), &userInfo)
	if err != nil {
		userInfo = nil
	}
	return userInfo, nil
}
