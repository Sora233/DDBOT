package concern_manager

import (
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	localutils "github.com/Sora233/Sora233-MiraiGo/utils"
	"github.com/Sora233/sliceutil"
	"github.com/tidwall/buntdb"
	"strings"
	"time"
)

var logger = utils.GetModuleLogger("concern_manager")

type StateManager struct {
	*localdb.ShortCut
	KeySet
	emitChan  chan interface{}
	emitQueue *localutils.EmitQueue
}

func (c *StateManager) CheckGroupConcern(groupCode int64, id interface{}, ctype concern.Type) error {
	return c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.GroupConcernStateKey(groupCode, id))
		if err == nil {
			if concern.FromString(val).ContainAll(ctype) {
				return ErrAlreadyExists
			}
		}
		return nil
	})
}

func (c *StateManager) CheckConcern(id interface{}, ctype concern.Type) error {
	return c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.ConcernStateKey(id))
		if err == nil {
			if concern.FromString(val).ContainAny(ctype) {
				return ErrAlreadyExists
			}
		}
		return nil
	})
}

func (c *StateManager) AddGroupConcern(groupCode int64, id interface{}, ctype concern.Type) (err error) {
	groupStateKey := c.GroupConcernStateKey(groupCode, id)
	stateKey := c.ConcernStateKey(id)
	err = c.upsertConcernType(groupStateKey, ctype)
	if err != nil {
		return err
	}
	if c.CheckConcern(id, concern.Empty) == nil {
		c.emitQueue.Add(id, time.Time{})
	}
	err = c.upsertConcernType(stateKey, ctype)
	return err
}

func (c *StateManager) Remove(groupCode int64, id interface{}, ctype concern.Type) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		groupStateKey := c.GroupConcernStateKey(groupCode, id)
		val, err := tx.Get(groupStateKey)
		if err != nil {
			return err
		}
		oldState := concern.FromString(val)
		newState := oldState.Remove(ctype)
		if newState == concern.Empty {
			_, err = tx.Delete(groupStateKey)
		} else {
			_, _, err = tx.Set(groupStateKey, newState.String(), nil)
		}
		if err != nil {
			return err
		}
		return nil
	})
}

func (c *StateManager) RemoveAll(groupCode int64) (err error) {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		var removeKey []string
		var iterErr error
		iterErr = tx.Ascend(c.GroupConcernStateKey(groupCode), func(key, value string) bool {
			removeKey = append(removeKey, key)
			return true
		})
		if iterErr != nil {
			return iterErr
		}
		for _, key := range removeKey {
			tx.Delete(key)
		}
		tx.DropIndex(c.GroupConcernStateKey(groupCode))
		return nil
	})
}

// GetGroupConcern return the concern.Type in specific group for a id
func (c *StateManager) GetGroupConcern(groupCode int64, id interface{}) (result concern.Type, err error) {
	err = c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.GroupConcernStateKey(groupCode, id))
		if err != nil {
			return err
		}
		result = concern.FromString(val)
		return nil
	})
	return result, err
}

// GetConcern return the concern.Type combined from all group for a id
func (c *StateManager) GetConcern(id interface{}) (result concern.Type, err error) {
	err = c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.ConcernStateKey(id))
		if err != nil {
			return err
		}
		result = concern.FromString(val)
		return nil
	})
	return result, err
}

func (c *StateManager) List(filter func(groupCode int64, id interface{}, p concern.Type) bool) (idGroups []int64, ids []interface{}, idTypes []concern.Type, err error) {
	err = c.RTxCover(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.GroupConcernStateKey(), func(key, value string) bool {
			var groupCode int64
			var id interface{}
			groupCode, id, iterErr = c.ParseGroupConcernStateKey(key)
			if iterErr != nil {
				return false
			}
			ctype := concern.FromString(value)
			if ctype == concern.Empty {
				return true
			}
			if filter(groupCode, id, ctype) == true {
				idGroups = append(idGroups, groupCode)
				ids = append(ids, id)
				idTypes = append(idTypes, ctype)
			}
			return true
		})
		if err != nil {
			return err
		}
		if iterErr != nil {
			return iterErr
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return
}

func (c *StateManager) ListByGroup(groupCode int64, filter func(id interface{}, p concern.Type) bool) (ids []interface{}, idTypes []concern.Type, err error) {
	err = c.RTxCover(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.GroupConcernStateKey(groupCode), func(key, value string) bool {
			var id interface{}
			_, id, iterErr = c.ParseGroupConcernStateKey(key)
			if iterErr != nil {
				return false
			}
			ctype := concern.FromString(value)
			if filter(id, ctype) == true {
				ids = append(ids, id)
				idTypes = append(idTypes, ctype)
			}
			return true
		})
		if err != nil {
			return err
		}
		if iterErr != nil {
			return iterErr
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return
}

func (c *StateManager) GroupTypeById(ids []interface{}, types []concern.Type) ([]interface{}, []concern.Type, error) {
	if len(ids) != len(types) {
		return nil, nil, ErrLengthMismatch
	}
	var (
		idTypeMap  = make(map[interface{}]concern.Type)
		result     []interface{}
		resultType []concern.Type
	)
	for index := range ids {
		id := ids[index]
		t := types[index]
		if old, found := idTypeMap[id]; found {
			idTypeMap[id] = old.Add(t)
		} else {
			idTypeMap[id] = t
		}
	}

	for id, t := range idTypeMap {
		result = append(result, id)
		resultType = append(resultType, t)
	}
	return result, resultType, nil
}

func (c *StateManager) FreshCheck(id interface{}, setTTL bool) (result bool, err error) {
	err = c.RWTxCover(func(tx *buntdb.Tx) error {
		freshKey := c.FreshKey(id)
		_, err := tx.Get(freshKey)
		if err == buntdb.ErrNotFound {
			result = true
			if setTTL {
				_, _, err = tx.Set(freshKey, "", localdb.ExpireOption(time.Minute))
				if err != nil {
					return err
				}
			}
		} else if err != nil {
			return err
		}
		return nil
	})
	return
}

func (c *StateManager) FreshIndex() {
	db, err := localdb.GetClient()
	if err != nil {
		return
	}
	for _, groupInfo := range miraiBot.Instance.GroupList {
		db.CreateIndex(c.GroupConcernStateKey(groupInfo.Code), c.GroupConcernStateKey(groupInfo.Code, "*"), buntdb.IndexString)
	}
}

func (c *StateManager) FreshAll() {
	miraiBot.Instance.ReloadFriendList()
	miraiBot.Instance.ReloadGroupList()
	db, err := localdb.GetClient()
	if err != nil {
		return
	}
	allIndex, err := db.Indexes()
	if err != nil {
		return
	}
	for _, index := range allIndex {
		if strings.HasPrefix(index, c.GroupConcernStateKey()+":") {
			db.DropIndex(index)
		}
	}

	c.FreshIndex()

	var groupCodes []int64
	for _, groupInfo := range miraiBot.Instance.GroupList {
		groupCodes = append(groupCodes, groupInfo.Code)
	}
	var removeKey []string

	c.RTxCover(func(tx *buntdb.Tx) error {
		tx.Ascend(c.GroupConcernStateKey(), func(key, value string) bool {
			groupCode, _, err := c.ParseGroupConcernStateKey(key)
			if err != nil {
				removeKey = append(removeKey, key)
			} else if !sliceutil.Contains(groupCodes, groupCode) {
				removeKey = append(removeKey, key)
			}
			return true
		})
		return nil
	})
	c.RWTxCover(func(tx *buntdb.Tx) error {
		for _, key := range removeKey {
			tx.Delete(key)
		}
		return nil
	})
}

func (c *StateManager) upsertConcernType(key string, ctype concern.Type) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			tx.Set(key, ctype.String(), nil)
		} else if err == nil {
			newVal := concern.FromString(val).Add(ctype)
			tx.Set(key, newVal.String(), nil)
		} else {
			return err
		}
		return nil

	})
}

func (c *StateManager) Start() error {
	err := c.freshConcern()
	if err != nil {
		return err
	}
	_, ids, _, err := c.List(func(groupCode int64, id interface{}, p concern.Type) bool {
		return true
	})
	if err != nil {
		return err
	}
	idSet := make(map[interface{}]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	for id := range idSet {
		c.emitQueue.Add(id, time.Now())
	}
	return nil
}

func (c *StateManager) freshConcern() error {
	_, ids, types, err := c.List(func(groupCode int64, id interface{}, p concern.Type) bool {
		return true
	})
	if err != nil {
		return err
	}
	ids, types, err = c.GroupTypeById(ids, types)
	if err != nil {
		return err
	}

	all := make(map[string]bool)

	c.RTxCover(func(tx *buntdb.Tx) error {
		return tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
			all[key] = true
			return true
		})
	})

	c.RWTxCover(func(tx *buntdb.Tx) error {
		for index := range ids {
			id := ids[index]
			ctype := types[index]
			key := c.ConcernStateKey(id)
			if ctype == concern.Empty {
				tx.Delete(key)
			} else {
				tx.Set(key, ctype.String(), nil)
			}
			delete(all, key)
		}
		return nil
	})

	c.RWTxCover(func(tx *buntdb.Tx) error {
		for key := range all {
			tx.Delete(key)
		}
		return nil
	})
	return nil
}

func (c *StateManager) EmitFreshCore(name string, fresher func(ctype concern.Type, id interface{}) error) {
	for id := range c.emitChan {
		if ok, _ := c.FreshCheck(id, true); !ok {
			logger.WithField("id", id).WithField("result", ok).Trace("fresh check failed")
			continue
		}
		logger.WithField("id", id).Trace("fresh")
		ctype, err := c.GetConcern(id)
		if err != nil {
			logger.WithField("id", id).Errorf("get concern failed %v", err)
			continue
		}
		if err := fresher(ctype, id); err != nil {
			logger.WithField("id", id).WithField("name", name).Errorf("fresher error %v", err)
		}
	}
}

func NewStateManager(keySet KeySet) *StateManager {
	sm := &StateManager{
		emitChan: make(chan interface{}),
		KeySet:   keySet,
	}
	sm.emitQueue = localutils.NewEmitQueue(sm.emitChan, config.GlobalConfig.GetDuration("concern.emitInterval"))
	return sm
}
