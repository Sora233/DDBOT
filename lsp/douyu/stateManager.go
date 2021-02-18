package douyu

import (
	"encoding/json"
	"errors"
	"fmt"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/tidwall/buntdb"
	"time"
)

type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (c *StateManager) GetLiveInfo(id int64) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}

	err := c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentLiveKey(id))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), liveInfo)
		if err != nil {
			fmt.Println(val)
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

func (c *StateManager) Start() error {
	db, err := localdb.GetClient()
	if err == nil {
		db.CreateIndex(c.GroupConcernStateKey(), c.GroupConcernStateKey("*"), buntdb.IndexString)
		db.CreateIndex(c.CurrentLiveKey(), c.CurrentLiveKey("*"), buntdb.IndexString)
		db.CreateIndex(c.FreshKey(), c.FreshKey("*"), buntdb.IndexString)
		db.CreateIndex(c.ConcernStateKey(), c.ConcernStateKey("*"), buntdb.IndexBinary)
	}
	return c.StateManager.Start()
}

func NewStateManager() *StateManager {
	sm := &StateManager{}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet())
	return sm
}
