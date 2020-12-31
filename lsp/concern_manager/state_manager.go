package concern_manager

import (
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/forestgiant/sliceutil"
	"github.com/tidwall/buntdb"
	"math/rand"
	"strings"
	"time"
)

type StateManager struct {
	KeySet
}

func (c *StateManager) Check(groupCode int64, id int64, ctype concern.Type) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.ConcernStateKey(groupCode, id))
		if err == nil {
			if concern.FromString(val).ContainAll(ctype) {
				return ErrAlreadyExists
			}
		}
		return nil
	})
	return err
}

func (c *StateManager) Add(groupCode int64, id int64, ctype concern.Type) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		stateKey := c.ConcernStateKey(groupCode, id)
		val, err := tx.Get(stateKey)
		if err == buntdb.ErrNotFound {
			tx.Set(stateKey, ctype.String(), nil)
		} else if err == nil {
			newVal := concern.FromString(val).Add(ctype)
			tx.Set(stateKey, newVal.String(), nil)
		} else {
			return err
		}
		return nil
	})
	return err
}

func (c *StateManager) Remove(groupCode int64, id int64, ctype concern.Type) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		stateKey := c.ConcernStateKey(groupCode, id)
		val, err := tx.Get(stateKey)
		if err != nil {
			return err
		}
		oldState := concern.FromString(val)
		newState := oldState.Remove(ctype)
		_, _, err = tx.Set(stateKey, newState.String(), nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *StateManager) RemoveAll(groupCode int64) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		var removeKey []string
		var iterErr error
		iterErr = tx.Ascend(c.ConcernStateKey(groupCode), func(key, value string) bool {
			removeKey = append(removeKey, key)
			return true
		})
		if iterErr != nil {
			return iterErr
		}
		for _, key := range removeKey {
			tx.Delete(key)
		}
		tx.DropIndex(c.ConcernStateKey(groupCode))
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *StateManager) Get(groupCode int64, id int64) (concern.Type, error) {
	var result concern.Type
	db, err := localdb.GetClient()
	if err != nil {
		return result, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.ConcernStateKey(groupCode, id))
		if err != nil {
			return err
		}
		result = concern.FromString(val)
		return nil
	})
	return result, err
}

func (c *StateManager) ListId(filter func(groupCode int64, id int64, p concern.Type) bool) (idGroups []int64, ids []int64, idTypes []concern.Type, err error) {
	var db *buntdb.DB
	db, err = localdb.GetClient()
	if err != nil {
		return
	}
	err = db.View(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
			var groupCode, id int64
			groupCode, id, iterErr = c.ParseConcernStateKey(key)
			if iterErr != nil {
				return false
			}
			ctype := concern.FromString(value)
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

func (c *StateManager) ListIdByGroup(groupCode int64, filter func(id int64, p concern.Type) bool) (ids []int64, idTypes []concern.Type, err error) {
	var db *buntdb.DB
	db, err = localdb.GetClient()
	if err != nil {
		return
	}
	err = db.View(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.ConcernStateKey(groupCode), func(key, value string) bool {
			var id int64
			_, id, iterErr = c.ParseConcernStateKey(key)
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

func (c *StateManager) GroupTypeById(ids []int64, types []concern.Type) ([]int64, []concern.Type, error) {
	if len(ids) != len(types) {
		return nil, nil, ErrLengthMismatch
	}
	var (
		idTypeMap = make(map[int64]concern.Type)

		result     []int64
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

func (c *StateManager) FreshCheck(id int64, setTTL bool) (result bool, err error) {
	db, err := localdb.GetClient()
	if err != nil {
		return false, err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		freshKey := c.FreshKey(id)
		_, err := tx.Get(freshKey)
		if err == buntdb.ErrNotFound {
			result = true
			if setTTL {
				ttl := time.Second*30 + time.Duration(rand.Intn(30))*time.Second
				_, _, err = tx.Set(freshKey, "", &buntdb.SetOptions{Expires: true, TTL: ttl})
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
		db.CreateIndex(c.ConcernStateKey(groupInfo.Code), c.ConcernStateKey(groupInfo.Code, "*"), buntdb.IndexString)
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
		if strings.HasPrefix(index, c.ConcernStateKey()+":") {
			db.DropIndex(index)
		}
	}

	c.FreshIndex()

	var groupCodes []int64
	for _, groupInfo := range miraiBot.Instance.GroupList {
		groupCodes = append(groupCodes, groupInfo.Code)
	}
	var removeKey []string
	db.View(func(tx *buntdb.Tx) error {
		tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
			groupCode, _, err := c.ParseConcernStateKey(key)
			if err != nil {
				removeKey = append(removeKey, key)
			} else if !sliceutil.Contains(groupCodes, groupCode) {
				removeKey = append(removeKey, key)
			}
			return true
		})
		return nil
	})
	db.Update(func(tx *buntdb.Tx) error {
		for _, key := range removeKey {
			tx.Delete(key)
		}
		return nil
	})
}

func NewStateManager(keySet KeySet) *StateManager {
	return &StateManager{
		keySet,
	}
}
