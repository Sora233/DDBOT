package lsp

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/tidwall/buntdb"

	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
)

type KeySet struct{}

func (KeySet) GroupMessageImageKey(keys ...interface{}) string {
	return localdb.GroupMessageImageKey(keys...)
}

func (KeySet) GroupMuteKey(keys ...interface{}) string {
	return localdb.GroupMuteKey(keys...)
}

func (KeySet) ModeKey() string {
	return localdb.ModeKey()
}

func (KeySet) NewFriendRequestKey(keys ...interface{}) string {
	return localdb.NewFriendRequestKey(keys...)
}

func (KeySet) GroupInvitedKey(keys ...interface{}) string {
	return localdb.GroupInvitedKey(keys...)
}

type StateManager struct {
	*localdb.ShortCut
	KeySet
}

func (s *StateManager) Muted(groupCode uint32, uin uint32, t uint32) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		key := s.GroupMuteKey(groupCode, uin)
		if t == 0 {
			_, err = s.Delete(key)
		} else if t == math.MaxUint32 {
			// 开启全体禁言
			err = s.Set(key, "")
		} else { // t > 0
			err = s.Set(key, "", localdb.SetExpireOpt(time.Second*time.Duration(t)))
		}
		return err
	})
}

func (s *StateManager) IsMuted(groupCode, uin uint32) bool {
	return s.Exist(s.GroupMuteKey(groupCode, uin))
}

func (s *StateManager) SaveGroupInvitor(groupCode uint32, uin uint32) error {
	err := s.SetUint32(localdb.GroupInvitorKey(groupCode), uin, localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return localdb.ErrKeyExist
	}
	return err
}

func (s *StateManager) PopGroupInvitor(groupCode uint32) (uint32, error) {
	return s.DeleteUint32(localdb.GroupInvitorKey(groupCode))
}

func (s *StateManager) FreshIndex() {
	for _, pattern := range []localdb.KeyPatternFunc{
		s.NewFriendRequestKey, s.GroupInvitedKey,
	} {
		s.CreatePatternIndex(pattern, nil)
	}
}

type Mode string

const (
	PublicMode  Mode = "public"
	PrivateMode Mode = "private"
	ProtectMode Mode = "protect"
)

func (s *StateManager) SetMode(mode Mode) error {
	if mode != PublicMode && mode != PrivateMode && mode != ProtectMode {
		return fmt.Errorf("未知模式【%v】", mode)
	}
	return s.Set(s.ModeKey(), string(mode))
}

func (s *StateManager) IsMode(mode Mode) bool {
	return s.GetCurrentMode() == mode
}

func (s *StateManager) IsPublicMode() bool {
	return s.IsMode(PublicMode)
}

func (s *StateManager) IsPrivateMode() bool {
	return s.IsMode(PrivateMode)
}

func (s *StateManager) IsProtectMode() bool {
	return s.IsMode(ProtectMode)
}

func (s *StateManager) GetCurrentMode() Mode {
	var result Mode
	val, err := s.Get(s.ModeKey())
	if err != nil {
		result = PublicMode
	}
	result = Mode(val)
	if result != PublicMode && result != PrivateMode && result != ProtectMode {
		result = PublicMode
	}
	return result
}

func (s *StateManager) SaveGroupInvitedRequest(request *event.GroupInvite) error {
	return s.saveRequest(strconv.FormatUint(request.RequestSeq, 10), request, s.GroupInvitedKey)
}

func (s *StateManager) SaveNewFriendRequest(request *event.NewFriendRequest) error {
	return s.saveRequest(request.SourceUID, request, s.NewFriendRequestKey)
}

func (s *StateManager) ListNewFriendRequest() (results []*event.NewFriendRequest, err error) {
	err = s.RCoverTx(func(tx *buntdb.Tx) error {
		var (
			iterErr, err error
		)
		err = tx.Ascend(s.NewFriendRequestKey(), func(key, value string) bool {
			var item = new(event.NewFriendRequest)
			iterErr = s.GetJson(key, &item)
			if iterErr == nil {
				results = append(results, item)
				return true
			}
			return false
		})
		if err != nil {
			return nil
		}
		if iterErr != nil {
			return iterErr
		}
		return nil
	})
	return
}

func (s *StateManager) ListGroupInvitedRequest() (results []*event.GroupInvite, err error) {
	err = s.RCoverTx(func(tx *buntdb.Tx) error {
		var (
			iterErr, err error
		)
		err = tx.Ascend(s.GroupInvitedKey(), func(key, value string) bool {
			var item = new(event.GroupInvite)
			iterErr = s.GetJson(key, &item)
			if iterErr == nil {
				results = append(results, item)
				return true
			}
			return false
		})
		if err != nil {
			return nil
		}
		if iterErr != nil {
			return iterErr
		}
		return nil
	})
	return
}

func (s *StateManager) DeleteNewFriendRequest(requestId string) (err error) {
	_, err = s.Delete(s.NewFriendRequestKey(requestId))
	return
}

func (s *StateManager) DeleteGroupInvitedRequest(requestId uint64) (err error) {
	_, err = s.Delete(s.GroupInvitedKey(requestId))
	return
}

func (s *StateManager) GetNewFriendRequest(requestId string) (result *event.NewFriendRequest, err error) {
	err = s.getRequest(requestId, &result, s.NewFriendRequestKey)
	return
}

func (s *StateManager) GetGroupInvitedRequest(requestId uint64) (result *event.GroupInvite, err error) {
	err = s.getRequest(strconv.FormatUint(requestId, 10), &result, s.GroupInvitedKey)
	return
}

func (s *StateManager) saveRequest(requestId string, request interface{}, keyFunc localdb.KeyPatternFunc) error {
	return s.SetJson(keyFunc(requestId), request)
}

func (s *StateManager) getRequest(requestId string, request interface{}, keyFunc localdb.KeyPatternFunc) error {
	return s.GetJson(keyFunc(requestId), request)
}

func NewStateManager() *StateManager {
	return &StateManager{
		KeySet: KeySet{},
	}
}
