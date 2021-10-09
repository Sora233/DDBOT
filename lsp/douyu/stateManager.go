package douyu

import (
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

func (c *StateManager) GetLiveInfo(id int64) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}
	err := c.JsonGet(c.CurrentLiveKey(id), liveInfo)
	if err != nil {
		logger.Errorf("JsonGet live info failed")
		return nil, err
	}
	return liveInfo, nil
}

func (c *StateManager) AddLiveInfo(liveInfo *LiveInfo) error {
	if liveInfo == nil {
		return errors.New("nil LiveInfo")
	}

	return c.RWCoverTx(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(c.CurrentLiveKey(liveInfo.RoomId), liveInfo.ToString(), localdb.ExpireOption(time.Hour*24*7))
		return err
	})
}

func (c *StateManager) Start() error {
	for _, pattern := range []localdb.KeyPatternFunc{c.GroupConcernStateKey, c.CurrentLiveKey, c.FreshKey} {
		c.CreatePatternIndex(pattern, nil)
	}
	return c.StateManager.Start()
}

func NewStateManager() *StateManager {
	sm := &StateManager{}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet(), true)
	return sm
}
