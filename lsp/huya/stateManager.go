package huya

import (
	"encoding/json"
	"errors"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/tidwall/buntdb"
	"time"
)

type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (c *StateManager) GetLiveInfo(id string) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}

	err := c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentLiveKey(id))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), liveInfo)
		if err != nil {
			logger.WithField("id", id).WithField("dbval", val).Errorf("Unmarshal error %v", err)
			return err
		}
		return nil

	})
	if err != nil {
		return nil, err
	}
	return liveInfo, nil
}

func (c *StateManager) AddLiveInfo(liveInfo *LiveInfo) error {
	if liveInfo == nil {
		return errors.New("nil LiveInfo")
	}

	return c.RWTxCover(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(c.CurrentLiveKey(liveInfo.RoomId), liveInfo.ToString(), localdb.ExpireOption(time.Hour*24))
		return err

	})
}

// ?为什么没有泛型?

func NewStateManager() *StateManager {
	sm := &StateManager{}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet(), true)
	return sm
}
