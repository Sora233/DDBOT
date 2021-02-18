package lsp

import (
	"github.com/Mrs4s/MiraiGo/message"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/utils"
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
		case *message.ImageElement:
			if i.Url != "" {
				urls = append(urls, i.Url)
			}
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
		logger.WithField("group_code", groupCode).
			WithField("message_id", messageID).
			WithField("urls", urls).Debug("save image")
	} else {
		return nil
	}
	return s.RWTxCover(func(tx *buntdb.Tx) error {
		key := s.GroupMessageImageKey(groupCode, messageID)
		_, _, err := tx.Set(key, strings.Join(urls, " "), localdb.ExpireOption(time.Hour*8))
		return err
	})
}

func (s *StateManager) GetMessageImageUrl(groupCode int64, messageID int32) []string {
	var result []string
	s.RTxCover(func(tx *buntdb.Tx) error {
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
	return s.RWTxCover(func(tx *buntdb.Tx) error {
		key := s.GroupMuteKey(groupCode, uin)
		if t == 0 {
			_, err := tx.Delete(key)
			return err
		} else {
			_, _, err := tx.Set(key, "", localdb.ExpireOption(time.Second*time.Duration(t)))
			return err
		}
	})
}

func (s *StateManager) IsMuted(groupCode int64, uin int64) bool {
	var result = true
	err := s.RTxCover(func(tx *buntdb.Tx) error {
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
	return s.RWTxCover(func(tx *buntdb.Tx) error {
		key := localdb.GroupInvitorKey(groupCode)
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			_, _, err := tx.Set(key, strconv.FormatInt(uin, 10), nil)
			return err
		} else if err != nil {
			return err
		} else {
			return localdb.ErrKeyExist
		}
	})
}

func (s *StateManager) GetGroupInvitor(groupCode int64) (target int64, err error) {
	err = s.RWTxCover(func(tx *buntdb.Tx) error {
		key := localdb.GroupInvitorKey(groupCode)
		invitor, err := tx.Delete(key)
		if err != nil {
			return err
		} else {
			target, err = strconv.ParseInt(invitor, 64, 10)
			if err != nil {
				return err
			}
			return nil
		}
	})
	return
}

func (s *StateManager) FreshIndex() {
	db := localdb.MustGetClient()
	db.CreateIndex(s.GroupMessageImageKey(), s.GroupMessageImageKey("*"), buntdb.IndexString)
	db.CreateIndex(s.GroupMuteKey(), s.GroupMuteKey("*"), buntdb.IndexString)
}

func NewStateManager() *StateManager {
	return &StateManager{
		KeySet: KeySet{},
	}
}
