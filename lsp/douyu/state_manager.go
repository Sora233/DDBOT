package douyu

import (
	"encoding/json"
	"errors"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/tidwall/buntdb"
)

type StateManager struct {
	*concern_manager.StateManager
}

func (c *StateManager) GetLiveInfo(id int64) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentLiveKey(id))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), liveInfo)
		if err != nil {
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
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(c.CurrentLiveKey(liveInfo.RoomId), liveInfo.ToString(), nil)
		return err
	})
	return err
}

func NewStateManager() *StateManager {
	sm := &StateManager{}
	sm.StateManager = concern_manager.NewStateManager(NewKeySet())
	return sm
}
