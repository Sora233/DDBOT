package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/utils"
	"github.com/tidwall/buntdb"
	"strconv"
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

//
//func (s *StateManager) SaveGroupMessage(msg *message.GroupMessage) error {
//
//}

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
	if len(urls) > 0 {
		//logger.WithFields(utils.GroupLogFields(groupCode)).
		//	WithField("message_id", messageID).
		//	WithField("urls", urls).Trace("save image")
	} else {
		return nil
	}
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := s.GroupMessageImageKey(groupCode, messageID)
		_, _, err := tx.Set(key, strings.Join(urls, " "), localdb.ExpireOption(time.Hour*8))
		return err
	})
}

func (s *StateManager) GetMessageImageUrl(groupCode int64, messageID int32) []string {
	var result []string
	s.RCoverTx(func(tx *buntdb.Tx) error {
		key := s.GroupMessageImageKey(groupCode, messageID)
		val, err := tx.Get(key)
		if err == nil {
			result = strings.Split(val, " ")
		}
		return err
	})
	return result
}

func (s *StateManager) Muted(groupCode int64, uin int64, t int32) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		key := s.GroupMuteKey(groupCode, uin)
		if t == 0 {
			_, err = tx.Delete(key)
			return err
		} else if t < 0 {
			if uin == 0 {
				// 开启全体禁言
				_, _, err = tx.Set(key, "", nil)
			} else {
				// 可能有吗？
				_, err = tx.Delete(key)
			}
		} else { // t > 0
			_, _, err = tx.Set(key, "", localdb.ExpireOption(time.Second*time.Duration(t)))
		}
		return err
	})
}

func (s *StateManager) IsMuted(groupCode int64, uin int64) bool {
	var result = true
	err := s.RCoverTx(func(tx *buntdb.Tx) error {
		key := s.GroupMuteKey(groupCode, uin)
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			result = false
			return nil
		} else {
			return err
		}
	})
	if err != nil {
		result = false
	}
	return result
}

func (s *StateManager) SaveGroupInvitor(groupCode int64, uin int64) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := localdb.GroupInvitorKey(groupCode)
		_, replaced, err := tx.Set(key, strconv.FormatInt(uin, 10), nil)
		if err != nil {
			return err
		}
		if replaced {
			return localdb.ErrKeyExist
		}
		return nil
	})
}

func (s *StateManager) PopGroupInvitor(groupCode int64) (target int64, err error) {
	err = s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := localdb.GroupInvitorKey(groupCode)
		invitor, err := tx.Delete(key)
		if err != nil {
			return err
		}
		target, err = strconv.ParseInt(invitor, 10, 64)
		return err
	})
	return
}

func (s *StateManager) FreshIndex() {
	db := localdb.MustGetClient()
	db.CreateIndex(s.GroupMessageImageKey(), s.GroupMessageImageKey("*"), buntdb.IndexString)
	db.CreateIndex(s.GroupMuteKey(), s.GroupMuteKey("*"), buntdb.IndexString)
	db.CreateIndex(s.NewFriendRequestKey(), s.NewFriendRequestKey("*"), buntdb.IndexString)
	db.CreateIndex(s.GroupInvitedKey(), s.GroupInvitedKey("*"), buntdb.IndexString)
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
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := s.ModeKey()
		_, _, err := tx.Set(key, string(mode), nil)
		return err
	})
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
	err := s.RCoverTx(func(tx *buntdb.Tx) error {
		key := s.ModeKey()
		val, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			result = PublicMode
			return nil
		} else if err != nil {
			return err
		}
		result = Mode(val)
		return nil
	})
	if err != nil {
		result = PublicMode
	}
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
			iterErr = s.JsonGet(key, &item)
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
			iterErr = s.JsonGet(key, &item)
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
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := s.NewFriendRequestKey(requestId)
		_, err := tx.Delete(key)
		return err
	})
}

func (s *StateManager) DeleteGroupInvitedRequest(requestId int64) (err error) {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := s.GroupInvitedKey(requestId)
		_, err := tx.Delete(key)
		return err
	})
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
	return s.JsonSave(keyFunc(requestId), request)
}

func (s *StateManager) getRequest(requestId int64, request interface{}, keyFunc localdb.KeyPatternFunc) error {
	return s.JsonGet(keyFunc(requestId), request)
}

func NewStateManager() *StateManager {
	return &StateManager{
		KeySet: KeySet{},
	}
}
