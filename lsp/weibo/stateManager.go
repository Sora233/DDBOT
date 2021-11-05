package weibo

import (
	"errors"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/tidwall/buntdb"
	"time"
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
		err = s.JsonSave(s.UserInfoKey(info.Uid), info.UserInfo, true)
		if err != nil {
			return err
		}
		return s.JsonSave(s.NewsInfoKey(info.Uid), info, true)
	})
}

func (s *StateManager) GetNewsInfo(uid int64) (*NewsInfo, error) {
	var newsInfo *NewsInfo
	err := s.JsonGet(s.NewsInfoKey(uid), &newsInfo)
	if err != nil {
		return nil, err
	}
	return newsInfo, nil
}

func (s *StateManager) MarkMblogId(mblogId string) (replaced bool, err error) {
	err = s.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		_, replaced, err = tx.Set(s.MarkMblogIdKey(mblogId), "", localdb.ExpireOption(time.Hour*120))
		return err
	})
	return
}
