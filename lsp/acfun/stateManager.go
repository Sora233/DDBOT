package acfun

import (
	"errors"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/tidwall/buntdb"
)

type StateManager struct {
	*concern.StateManager
	extraKey
}

func (s *StateManager) GetGroupConcernConfig(groupCode int64, id interface{}) (concernConfig concern.IConfig) {
	return NewGroupConcernConfig(s.StateManager.GetGroupConcernConfig(groupCode, id))
}

func NewStateManager(notify chan<- concern.Notify) *StateManager {
	return &StateManager{
		StateManager: concern.NewStateManagerWithInt64ID("Acfun", notify),
	}
}

func (s *StateManager) GetUserInfo(uid int64) (*UserInfo, error) {
	var userInfo *UserInfo
	err := s.GetJson(s.UserInfoKey(uid), &userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (s *StateManager) AddUserInfo(info *UserInfo) error {
	if info == nil {
		return errors.New("<nil userinfo>")
	}
	return s.SetJson(s.UserInfoKey(info.Uid), info)
}

func (s *StateManager) AddLiveInfo(info *LiveInfo) error {
	return s.RWCover(func() error {
		err := s.SetJson(s.UserInfoKey(info.Uid), info.UserInfo)
		if err != nil {
			return err
		}
		err = s.SetJson(s.LiveInfoKey(info.Uid), info)
		return err
	})
}

func (s *StateManager) GetLiveInfo(uid int64) (*LiveInfo, error) {
	var liveInfo *LiveInfo
	err := s.GetJson(s.LiveInfoKey(uid), &liveInfo)
	if err != nil {
		return nil, err
	}
	return liveInfo, nil
}

func (s *StateManager) DeleteLiveInfo(uid int64) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(s.LiveInfoKey(uid))
		return err
	})
}

func (s *StateManager) IncNotLiveCount(uid int64) int64 {
	result, err := s.SeqNext(s.NotLiveKey(uid))
	if err != nil {
		result = 0
	}
	return result
}

func (s *StateManager) ClearNotLiveCount(uid int64) error {
	_, err := s.Delete(s.NotLiveKey(uid), localdb.IgnoreNotFoundOpt())
	return err
}

func (s *StateManager) SetUidFirstTimestampIfNotExist(uid int64, timestamp int64) error {
	return s.SetInt64(s.UidFirstTimestamp(uid), timestamp, localdb.SetNoOverWriteOpt())
}

func (s *StateManager) GetUidFirstTimestamp(uid int64) (timestamp int64, err error) {
	timestamp, err = s.GetInt64(s.UidFirstTimestamp(uid))
	if err != nil {
		timestamp = 0
	}
	return
}
