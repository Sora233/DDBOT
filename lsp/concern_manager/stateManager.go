package concern_manager

import (
	"errors"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/concern"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/sliceutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"strings"
	"sync"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var logger = utils.GetModuleLogger("concern_manager")
var ErrEmitQNotInit = errors.New("emit queue not enabled")

type StateManager struct {
	*localdb.ShortCut
	KeySet
	emitChan  chan *localutils.EmitE
	emitQueue *localutils.EmitQueue
	useEmit   bool
	stop      chan interface{}
	wg        sync.WaitGroup
}

// GetGroupConcernConfig always return non-nil
func (c *StateManager) GetGroupConcernConfig(groupCode int64, id interface{}) (concernConfig *GroupConcernConfig) {
	var err error
	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.GroupConcernConfigKey(groupCode, id))
		if err != nil {
			return err
		}
		concernConfig, err = NewGroupConcernConfigFromString(val)
		return err
	})
	if err != nil && err != buntdb.ErrNotFound {
		logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("id", id).Errorf("GetGroupConcernConfig error %v", err)
	}
	if concernConfig == nil {
		concernConfig = new(GroupConcernConfig)
	}
	return
}

// OperateGroupConcernConfig 在一个rw事务中获取GroupConcernConfig并交给函数，如果返回true，就保存GroupConcernConfig，否则就回滚。
func (c *StateManager) OperateGroupConcernConfig(groupCode int64, id interface{}, f func(concernConfig *GroupConcernConfig) bool) error {
	err := c.RWCoverTx(func(tx *buntdb.Tx) error {
		var concernConfig *GroupConcernConfig
		var configKey = c.GroupConcernConfigKey(groupCode, id)
		val, err := tx.Get(configKey)
		if err == nil {
			concernConfig, err = NewGroupConcernConfigFromString(val)
		} else if err == buntdb.ErrNotFound {
			concernConfig = new(GroupConcernConfig)
			err = nil
		}
		if err != nil {
			return err
		}
		if !f(concernConfig) {
			return localdb.ErrRollback
		}
		_, _, err = tx.Set(configKey, concernConfig.ToString(), nil)
		return err
	})
	return err
}

// CheckAndSetAtAllMark 检查@全体标记是否过期，已过期返回true，并重置标记，否则返回false。
// 因为@全体有次数限制，并且较为恼人，故设置标记，两次@全体之间必须有间隔。
func (c *StateManager) CheckAndSetAtAllMark(groupCode int64, id interface{}) (result bool) {
	err := c.RWCoverTx(func(tx *buntdb.Tx) error {
		key := c.GroupAtAllMarkKey(groupCode, id)
		_, replaced, err := tx.Set(key, "", localdb.ExpireOption(time.Hour*2))
		if err != nil {
			return err
		}
		if replaced {
			return localdb.ErrRollback
		}
		return nil
	})
	if err == nil {
		result = true
	}
	return
}

func (c *StateManager) CheckGroupConcern(groupCode int64, id interface{}, ctype concern.Type) error {
	return c.RCoverTx(func(tx *buntdb.Tx) error {
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
	state, err := c.GetConcern(id)
	if err != nil {
		return err
	}
	if state.ContainAll(ctype) {
		return ErrAlreadyExists
	}
	return nil
}

func (c *StateManager) AddGroupConcern(groupCode int64, id interface{}, ctype concern.Type) (newCtype concern.Type, err error) {
	groupStateKey := c.GroupConcernStateKey(groupCode, id)
	oldCtype, err := c.GetConcern(id)
	if err != nil {
		return concern.Empty, err
	}
	newCtype, err = c.upsertConcernType(groupStateKey, ctype)
	if err != nil {
		return concern.Empty, err
	}

	if c.useEmit && oldCtype.Empty() {
		for _, t := range ctype.Split() {
			c.emitQueue.Add(localutils.NewEmitE(id, t), time.Time{})
		}
	}
	return
}

func (c *StateManager) RemoveGroupConcern(groupCode int64, id interface{}, ctype concern.Type) (newCtype concern.Type, err error) {
	err = c.RWCoverTx(func(tx *buntdb.Tx) error {
		groupStateKey := c.GroupConcernStateKey(groupCode, id)
		val, err := tx.Get(groupStateKey)
		if err != nil {
			return err
		}
		oldState := concern.FromString(val)
		newCtype = oldState.Remove(ctype)
		if newCtype.Empty() {
			_, err = tx.Delete(groupStateKey)
		} else {
			_, _, err = tx.Set(groupStateKey, newCtype.String(), nil)
		}
		return err
	})
	return
}

func (c *StateManager) RemoveAllByGroupCode(groupCode int64) (keys []string, err error) {
	var indexKey = []string{
		c.GroupConcernStateKey(),
		c.GroupConcernConfigKey(),
	}
	var prefixKey = []string{
		c.GroupConcernStateKey(groupCode),
		c.GroupConcernConfigKey(groupCode),
	}
	return localdb.RemoveByPrefixAndIndex(prefixKey, indexKey)
}

func (c *StateManager) RemoveAllById(_id interface{}) (err error) {
	return c.RWCoverTx(func(tx *buntdb.Tx) error {
		var removeKey []string
		var iterErr error
		iterErr = tx.Ascend(c.GroupConcernStateKey(), func(key, value string) bool {
			var id interface{}
			_, id, iterErr = c.ParseGroupConcernStateKey(key)
			if id == _id {
				removeKey = append(removeKey, key)
			}
			return true
		})
		if iterErr != nil {
			return iterErr
		}
		for _, key := range removeKey {
			tx.Delete(key)
		}
		return nil
	})
}

// GetGroupConcern return the concern.Type in specific group for an id
func (c *StateManager) GetGroupConcern(groupCode int64, id interface{}) (result concern.Type, err error) {
	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.GroupConcernStateKey(groupCode, id))
		if err != nil {
			return err
		}
		result = concern.FromString(val)
		return nil
	})
	return result, err
}

// GetConcern return the concern.Type combined from all group for an id
func (c *StateManager) GetConcern(id interface{}) (result concern.Type, err error) {
	_, _, _, err = c.List(func(groupCode int64, _id interface{}, p concern.Type) bool {
		if id == _id {
			result = result.Add(p)
		}
		return true
	})
	return
}

func (c *StateManager) List(filter func(groupCode int64, id interface{}, p concern.Type) bool) (idGroups []int64, ids []interface{}, idTypes []concern.Type, err error) {
	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.GroupConcernStateKey(), func(key, value string) bool {
			var groupCode int64
			var id interface{}
			groupCode, id, iterErr = c.ParseGroupConcernStateKey(key)
			if iterErr != nil {
				return false
			}
			ctype := concern.FromString(value)
			if ctype.Empty() {
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
	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.GroupConcernStateKey(groupCode), func(key, value string) bool {
			var id interface{}
			_, id, iterErr = c.ParseGroupConcernStateKey(key)
			if iterErr != nil {
				return false
			}
			ctype := concern.FromString(value)
			if filter == nil || filter(id, ctype) == true {
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

func (c *StateManager) ListIds() (ids []interface{}, err error) {
	var idSet = make(map[interface{}]bool)
	_, _, _, err = c.List(func(groupCode int64, id interface{}, p concern.Type) bool {
		idSet[id] = true
		return true
	})
	if err != nil {
		return nil, err
	}
	for k := range idSet {
		ids = append(ids, k)
	}
	return ids, nil
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
	err = c.RWCoverTx(func(tx *buntdb.Tx) error {
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

func (c *StateManager) FreshIndex(groups ...int64) {
	for _, pattern := range []localdb.KeyPatternFunc{
		c.GroupConcernStateKey, c.GroupConcernConfigKey,
		c.GroupAtAllMarkKey, c.FreshKey} {
		c.CreatePatternIndex(pattern, nil)
	}
	var groupSet = make(map[int64]interface{})
	if len(groups) == 0 {
		for _, groupInfo := range miraiBot.Instance.GroupList {
			groupSet[groupInfo.Code] = struct{}{}
		}
	} else {
		for _, g := range groups {
			groupSet[g] = struct{}{}
		}
	}
	c.List(func(groupCode int64, id interface{}, p concern.Type) bool {
		groupSet[groupCode] = struct{}{}
		return true
	})
	for g := range groupSet {
		for _, pattern := range []localdb.KeyPatternFunc{
			c.GroupConcernStateKey, c.GroupConcernConfigKey,
			c.GroupAtAllMarkKey, c.FreshKey} {
			c.CreatePatternIndex(pattern, []interface{}{g})
		}
	}
}

func (c *StateManager) FreshAll() {
	if err := miraiBot.Instance.ReloadFriendList(); err != nil {
		logger.Errorf("ReloadFriendList error %v", err)
		return
	}
	if err := miraiBot.Instance.ReloadGroupList(); err != nil {
		logger.Errorf("ReloadGroupList error %v", err)
		return
	}
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
			if err := db.DropIndex(index); err != nil {
				logger.Errorf("DropIndex %v error %v", index, err)
			}
		}
	}

	c.FreshIndex()

	var groupCodes []int64
	for _, groupInfo := range miraiBot.Instance.GroupList {
		groupCodes = append(groupCodes, groupInfo.Code)
	}
	var removeKey []string

	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		return tx.Ascend(c.GroupConcernStateKey(), func(key, value string) bool {
			groupCode, _, err := c.ParseGroupConcernStateKey(key)
			if err != nil {
				removeKey = append(removeKey, key)
			} else if !sliceutil.Contains(groupCodes, groupCode) {
				removeKey = append(removeKey, key)
			}
			return true
		})
	})
	if err != nil {
		logger.Errorf("Ascend %v error %v", c.GroupConcernStateKey(), err)
		return
	}
	err = c.RWCoverTx(func(tx *buntdb.Tx) error {
		for _, key := range removeKey {
			if _, err := tx.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Errorf("tx delete error %v", err)
	}
}

func (c *StateManager) upsertConcernType(key string, ctype concern.Type) (newCtype concern.Type, err error) {
	err = c.RWCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			newCtype = ctype
			_, _, err = tx.Set(key, ctype.String(), nil)
		} else if err == nil {
			newCtype = concern.FromString(val).Add(ctype)
			_, _, err = tx.Set(key, newCtype.String(), nil)
		} else {
			return err
		}
		return err
	})
	return
}

func (c *StateManager) Stop() {
	if c.stop != nil {
		close(c.stop)
	}
	c.wg.Wait()
}

func (c *StateManager) Start() error {
	if c.useEmit {
		c.emitQueue.Start()
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
		for index := range ids {
			for _, t := range types[index].Split() {
				c.emitQueue.Add(localutils.NewEmitE(ids[index], t), time.Now())
			}
		}
	}
	return nil
}

func (c *StateManager) EmitFreshCore(name string, fresher func(ctype concern.Type, id interface{}) error) {
	if !c.useEmit {
		return
	}
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		select {
		case e := <-c.emitChan:
			id := e.Id
			if ok, _ := c.FreshCheck(id, true); !ok {
				logger.WithFields(logrus.Fields{
					"Id":     id,
					"Result": ok,
				}).Trace("fresh check failed")
				continue
			}
			logger.WithField("id", id).Trace("fresh")
			if err := fresher(e.Type, id); err != nil {
				logger.WithFields(logrus.Fields{
					"Id":   id,
					"Name": name,
				}).Errorf("fresher error %v", err)
			}
		case <-c.stop:
			return
		}
	}
}

func NewStateManager(keySet KeySet, useEmit bool) *StateManager {
	sm := &StateManager{
		emitChan: make(chan *localutils.EmitE),
		KeySet:   keySet,
		useEmit:  useEmit,
		stop:     make(chan interface{}),
	}
	if useEmit {
		var interval time.Duration
		if config.GlobalConfig == nil {
			interval = time.Second * 5
		} else {
			interval = config.GlobalConfig.GetDuration("concern.emitInterval")
		}
		sm.emitQueue = localutils.NewEmitQueue(sm.emitChan, interval)
	}
	return sm
}
