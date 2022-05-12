package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/utils"
	"github.com/tidwall/buntdb"
	"strings"
	"time"
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

func (s *StateManager) SaveMessageImageUrl(groupCode int64, messageID int32, msgs []message.IMessageElement) error {
	imgs := utils.MessageFilter(msgs, func(e message.IMessageElement) bool {
		return e.Type() == message.Image
	})
	var urls []string
	for _, img := range imgs {
		switch i := img.(type) {
		case *message.GroupImageElement:
			if i.Url != "" {
				urls = append(urls, i.Url)
			}
		case *message.FriendImageElement:
			if i.Url != "" {
				urls = append(urls, i.Url)
			}
		}
	}
	if len(urls) == 0 {
		return nil
	}
	return s.Set(s.GroupMessageImageKey(groupCode, messageID), strings.Join(urls, " "), localdb.SetExpireOpt(time.Hour*8))
}

func (s *StateManager) GetMessageImageUrl(groupCode int64, messageID int32) []string {
	var result []string
	val, err := s.Get(s.GroupMessageImageKey(groupCode, messageID))
	if err == nil {
		result = strings.Split(val, " ")
	}
	return result
}

func (s *StateManager) Muted(target mt.Target, uin int64, t int32) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		key := s.GroupMuteKey(target, uin)
		if t == 0 {
			_, err = s.Delete(key)
		} else if t < 0 {
			// 开启全体禁言
			err = s.Set(key, "")
		} else { // t > 0
			err = s.Set(key, "", localdb.SetExpireOpt(time.Second*time.Duration(t)))
		}
		return err
	})
}

func (s *StateManager) IsMuted(target mt.Target, uin int64) bool {
	if target.GetTargetType().IsGroup() {
		return s.Exist(s.GroupMuteKey(target, uin))
	}
	return false
}

func (s *StateManager) SaveGroupInvitor(groupCode int64, uin int64) error {
	err := s.SetInt64(localdb.GroupInvitorKey(groupCode), uin, localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return localdb.ErrKeyExist
	}
	return err
}

func (s *StateManager) PopGroupInvitor(groupCode int64) (int64, error) {
	return s.DeleteInt64(localdb.GroupInvitorKey(groupCode))
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

func (s *StateManager) SaveGroupInvitedRequest(request *client.GroupInvitedRequest) error {
	return s.saveRequest(request.RequestId, request, s.GroupInvitedKey)
}

func (s *StateManager) SaveNewFriendRequest(request *client.NewFriendRequest) error {
	return s.saveRequest(request.RequestId, request, s.NewFriendRequestKey)
}

func (s *StateManager) ListNewFriendRequest() (results []*client.NewFriendRequest, err error) {
	err = s.RCoverTx(func(tx *buntdb.Tx) error {
		var (
			iterErr, err error
		)
		err = tx.Ascend(s.NewFriendRequestKey(), func(key, value string) bool {
			var item = new(client.NewFriendRequest)
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

func (s *StateManager) ListGroupInvitedRequest() (results []*client.GroupInvitedRequest, err error) {
	err = s.RCoverTx(func(tx *buntdb.Tx) error {
		var (
			iterErr, err error
		)
		err = tx.Ascend(s.GroupInvitedKey(), func(key, value string) bool {
			var item = new(client.GroupInvitedRequest)
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

func (s *StateManager) DeleteNewFriendRequest(requestId int64) (err error) {
	_, err = s.Delete(s.NewFriendRequestKey(requestId))
	return
}

func (s *StateManager) DeleteGroupInvitedRequest(requestId int64) (err error) {
	_, err = s.Delete(s.GroupInvitedKey(requestId))
	return
}

func (s *StateManager) GetNewFriendRequest(requestId int64) (result *client.NewFriendRequest, err error) {
	err = s.getRequest(requestId, &result, s.NewFriendRequestKey)
	return
}

func (s *StateManager) GetGroupInvitedRequest(requestId int64) (result *client.GroupInvitedRequest, err error) {
	err = s.getRequest(requestId, &result, s.GroupInvitedKey)
	return
}

func (s *StateManager) saveRequest(requestId int64, request interface{}, keyFunc localdb.KeyPatternFunc) error {
	return s.SetJson(keyFunc(requestId), request)
}

func (s *StateManager) getRequest(requestId int64, request interface{}, keyFunc localdb.KeyPatternFunc) error {
	return s.GetJson(keyFunc(requestId), request)
}

func NewStateManager() *StateManager {
	return &StateManager{
		KeySet: KeySet{},
	}
}
