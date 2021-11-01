package concern

import (
	"errors"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("concern")
var ErrEmitQueueNotInit = errors.New("emit queue not enabled")

type IStateManager interface {
	Start() error
	Stop()

	GetGroupConcernConfig(groupCode int64, id interface{}) (concernConfig IConfig)
	OperateGroupConcernConfig(groupCode int64, id interface{}, f func(concernConfig IConfig) bool) error

	GetGroupConcern(groupCode int64, id interface{}) (result concern_type.Type, err error)
	GetConcern(id interface{}) (result concern_type.Type, err error)

	CheckAndSetAtAllMark(groupCode int64, id interface{}) (result bool)
	CheckGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) error
	CheckConcern(id interface{}, ctype concern_type.Type) error
	CheckFresh(id interface{}, setTTL bool) (result bool, err error)

	AddGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error)
	RemoveGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error)
	RemoveAllByGroupCode(groupCode int64) (keys []string, err error)
	RemoveAllById(_id interface{}) (err error)

	ListConcernState(filter func(groupCode int64, id interface{}, p concern_type.Type) bool) (idGroups []int64,
		ids []interface{}, idTypes []concern_type.Type, err error)
	GroupTypeById(ids []interface{}, types []concern_type.Type) ([]interface{}, []concern_type.Type, error)

	FreshIndex(groups ...int64)
	EmitFreshCore(name string, fresher func(ctype concern_type.Type, id interface{}) error)
}

type StateManager struct {
	*localdb.ShortCut
	KeySet
	emitChan  chan *localutils.EmitE
	emitQueue *localutils.EmitQueue
	useEmit   bool
	stop      chan interface{}
	wg        sync.WaitGroup
}

func (c *StateManager) getGroupConcernConfig(groupCode int64, id interface{}) (concernConfig *GroupConcernConfig) {
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

// GetGroupConcernConfig always return non-nil
func (c *StateManager) GetGroupConcernConfig(groupCode int64, id interface{}) IConfig {
	return c.getGroupConcernConfig(groupCode, id)
}

// OperateGroupConcernConfig 在一个rw事务中获取GroupConcernConfig并交给函数，如果返回true，就保存GroupConcernConfig，否则就回滚。
func (c *StateManager) OperateGroupConcernConfig(groupCode int64, id interface{}, f func(concernConfig IConfig) bool) error {
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

func (c *StateManager) CheckGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) error {
	return c.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.GroupConcernStateKey(groupCode, id))
		if err == nil {
			if concern_type.FromString(val).ContainAll(ctype) {
				return ErrAlreadyExists
			}
		}
		return nil
	})
}

func (c *StateManager) CheckConcern(id interface{}, ctype concern_type.Type) error {
	state, err := c.GetConcern(id)
	if err != nil {
		return err
	}
	if state.ContainAll(ctype) {
		return ErrAlreadyExists
	}
	return nil
}

func (c *StateManager) AddGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	groupStateKey := c.GroupConcernStateKey(groupCode, id)
	oldCtype, err := c.GetConcern(id)
	if err != nil {
		return concern_type.Empty, err
	}
	newCtype, err = c.upsertConcernType(groupStateKey, ctype)
	if err != nil {
		return concern_type.Empty, err
	}

	if c.useEmit && oldCtype.Empty() {
		for _, t := range ctype.Split() {
			c.emitQueue.Add(localutils.NewEmitE(id, t), time.Time{})
		}
	}
	return
}

func (c *StateManager) RemoveGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	err = c.RWCoverTx(func(tx *buntdb.Tx) error {
		groupStateKey := c.GroupConcernStateKey(groupCode, id)
		val, err := tx.Get(groupStateKey)
		if err != nil {
			return err
		}
		oldState := concern_type.FromString(val)
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

// GetGroupConcern return the concern_type.Type in specific group for an id
func (c *StateManager) GetGroupConcern(groupCode int64, id interface{}) (result concern_type.Type, err error) {
	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.GroupConcernStateKey(groupCode, id))
		if err != nil {
			return err
		}
		result = concern_type.FromString(val)
		return nil
	})
	return result, err
}

// GetConcern 查询一个id在所有group内的 concern_type.Type
func (c *StateManager) GetConcern(id interface{}) (result concern_type.Type, err error) {
	_, _, _, err = c.ListConcernState(func(groupCode int64, _id interface{}, p concern_type.Type) bool {
		if id == _id {
			result = result.Add(p)
		}
		return true
	})
	return
}

func (c *StateManager) ListConcernState(filter func(groupCode int64, id interface{}, p concern_type.Type) bool) (idGroups []int64, ids []interface{}, idTypes []concern_type.Type, err error) {
	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.GroupConcernStateKey(), func(key, value string) bool {
			var groupCode int64
			var id interface{}
			groupCode, id, iterErr = c.ParseGroupConcernStateKey(key)
			if iterErr != nil {
				return false
			}
			ctype := concern_type.FromString(value)
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

func (c *StateManager) GroupTypeById(ids []interface{}, types []concern_type.Type) ([]interface{}, []concern_type.Type, error) {
	if len(ids) != len(types) {
		return nil, nil, ErrLengthMismatch
	}
	var (
		idTypeMap  = make(map[interface{}]concern_type.Type)
		result     []interface{}
		resultType []concern_type.Type
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

func (c *StateManager) CheckFresh(id interface{}, setTTL bool) (result bool, err error) {
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
	c.ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
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

func (c *StateManager) upsertConcernType(key string, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	err = c.RWCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			newCtype = ctype
			_, _, err = tx.Set(key, ctype.String(), nil)
		} else if err == nil {
			newCtype = concern_type.FromString(val).Add(ctype)
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
		_, ids, types, err := c.ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
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

func (c *StateManager) EmitFreshCore(name string, fresher func(ctype concern_type.Type, id interface{}) error) {
	if !c.useEmit {
		return
	}
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		select {
		case e := <-c.emitChan:
			id := e.Id
			if ok, _ := c.CheckFresh(id, true); !ok {
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

func NewStateManagerWithCustomKey(keySet KeySet, useEmit bool) *StateManager {
	sm := &StateManager{
		emitChan: make(chan *localutils.EmitE),
		KeySet:   keySet,
		useEmit:  useEmit,
		stop:     make(chan interface{}),
	}
	if useEmit {
		var interval time.Duration
		if config.GlobalConfig != nil {
			interval = config.GlobalConfig.GetDuration("concern.emitInterval")
		}
		if interval == 0 {
			interval = time.Second * 5
		}
		sm.emitQueue = localutils.NewEmitQueue(sm.emitChan, interval)
	}
	return sm
}

func NewStateManagerWithStringID(name string, useEmit bool) *StateManager {
	return NewStateManagerWithCustomKey(NewPrefixKeySetWithStringID(name), useEmit)
}

func NewStateManagerWithInt64ID(name string, useEmit bool) *StateManager {
	return NewStateManagerWithCustomKey(NewPrefixKeySetWithInt64ID(name), useEmit)
}