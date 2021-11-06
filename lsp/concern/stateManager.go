package concern

import (
	"context"
	"errors"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("concern")
var ErrEmitQueueNotInit = errors.New("emit queue not enabled")

// NotifyGeneratorFunc 是 IStateManager.NotifyGenerator 函数的具体逻辑
// 它针对一组 groupCode 把 Event 转变成一组 Notify
//
// 使用 StateManager 时，在 StateManager.Start 之前，
// 必须使用 StateManager.UseNotifyGeneratorFunc 来指定一个 NotifyGeneratorFunc, 否则会发生 panic
type NotifyGeneratorFunc func(groupCode int64, event Event) []Notify

// DispatchFunc 是 IStateManager.Dispatch 函数的具体逻辑
// 它从event channel中获取 Event，把 Event 转变成（可能多个） Notify 并发送到notify channel
//
// StateManager 可以使用 StateManager.UseDispatchFunc 来指定一个 DispatchFunc
// StateManager 中有一个默认实现，适用于绝大多数情况，请参考 StateManager.DefaultDispatch
type DispatchFunc func(event <-chan Event, notify chan<- Notify)

// FreshFunc 是 IStateManager.Fresh 函数的具体逻辑，没有具体的限制
// 对于大多数网站来说，它的逻辑是访问网页获取数据，和和之前的数据对比，判断新数据，产生 Event 发送给 eventChan
//
// 使用 StateManager 时，在 StateManager.Start 之前，必须使用 StateManager.UseFreshFunc 来指定一个 FreshFunc, 否则会发生 panic
type FreshFunc func(ctx context.Context, eventChan chan<- Event)

type IStateManager interface {
	GetGroupConcernConfig(groupCode int64, id interface{}) (concernConfig IConfig)
	OperateGroupConcernConfig(groupCode int64, id interface{}, f func(concernConfig IConfig) bool) error

	GetGroupConcern(groupCode int64, id interface{}) (result concern_type.Type, err error)
	GetConcern(id interface{}) (result concern_type.Type, err error)

	CheckAndSetAtAllMark(groupCode int64, id interface{}) (result bool)
	CheckGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) error
	CheckConcern(id interface{}, ctype concern_type.Type) error

	AddGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error)
	RemoveGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error)
	RemoveAllByGroupCode(groupCode int64) (keys []string, err error)

	ListConcernState(filter func(groupCode int64, id interface{}, p concern_type.Type) bool) (idGroups []int64,
		ids []interface{}, idTypes []concern_type.Type, err error)
	GroupTypeById(ids []interface{}, types []concern_type.Type) ([]interface{}, []concern_type.Type, error)

	// NotifyGenerator 从 Event 产生多个 Notify
	NotifyGenerator(groupCode int64, event Event) []Notify
	// Fresh 是一个长生命周期的函数，它产生 Event
	Fresh(wg *sync.WaitGroup, eventChan chan<- Event)
	// Dispatch 是一个长生命周期的函数，它从event channel中获取 Event， 并产生 Notify 发送到notify channel
	Dispatch(wg *sync.WaitGroup, event <-chan Event, notify chan<- Notify)
}

// StateManager 定义了一些通用的状态行为，例如添加订阅，删除订阅，查询订阅，
// 还默认集成了一种定时刷新策略，“每隔几秒钟刷新一个id”（即：轮询）这个策略，开箱即用，如果不需要使用，也可以自定义策略。
// StateManager 是每个 Concern 之间隔离的，不同的订阅源必须持有不同的 StateManager，
// 操作一个 StateManager 时不会对其他 StateManager 内的数据产生影响，
// 读取当前 StateManager 内的订阅时，也不会获取到其他 StateManager 内的订阅。
//
// StateManager 通过 KeySet 来隔离存储的数据，创建 StateManager 时必须使用唯一的 KeySet，
// 请通过 NewStateManagerWithStringID / NewStateManagerWithInt64ID / NewStateManagerWithCustomKey 来创建 StateManager。
type StateManager struct {
	*localdb.ShortCut
	KeySet

	name                string
	eventChan           chan Event
	notifyChan          chan<- Notify
	emitChan            chan *localutils.EmitE
	emitQueue           *localutils.EmitQueue
	useEmit             bool
	ctx                 context.Context
	cancelCtx           context.CancelFunc
	freshWg             sync.WaitGroup
	wg                  sync.WaitGroup
	freshFunc           FreshFunc
	dispatchFunc        DispatchFunc
	notifyGeneratorFunc NotifyGeneratorFunc
}

func (c *StateManager) getGroupConcernConfig(groupCode int64, id interface{}) (concernConfig *GroupConcernConfig) {
	val, err := c.Get(c.GroupConcernConfigKey(groupCode, id), localdb.IgnoreNotFoundOpt())
	if err != nil {
		logger.WithFields(localutils.GroupLogFields(groupCode)).
			WithField("id", id).
			Errorf("GetGroupConcernConfig error %v", err)
	}
	if len(val) > 0 {
		concernConfig, err = NewGroupConcernConfigFromString(val)
		if err != nil {
			logger.WithFields(localutils.GroupLogFields(groupCode)).
				WithFields(logrus.Fields{"id": id, "val": val}).Errorf("NewGroupConcernConfigFromString error %v", err)
		}
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
		concernConfig := c.getGroupConcernConfig(groupCode, id)
		if !f(concernConfig) {
			return localdb.ErrRollback
		}
		return c.Set(c.GroupConcernConfigKey(groupCode, id), concernConfig.ToString())
	})
	return err
}

// CheckAndSetAtAllMark 检查@全体标记是否过期，未设置过或已过期返回true，并重置标记，否则返回false。
// 因为@全体有次数限制，并且较为恼人，故设置标记，两次@全体之间必须有间隔。
func (c *StateManager) CheckAndSetAtAllMark(groupCode int64, id interface{}) (result bool) {
	err := c.Set(c.GroupAtAllMarkKey(groupCode, id), "",
		localdb.SetExpireOpt(time.Hour*2), localdb.SetNoOverWriteOpt())
	return err == nil
}

// CheckGroupConcern 检查group是否已经添加过id的ctype订阅，如果添加过，返回 ErrAlreadyExists
func (c *StateManager) CheckGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) error {
	state, _ := c.GetGroupConcern(groupCode, id)
	if state.ContainAll(ctype) {
		return ErrAlreadyExists
	}
	return nil
}

// CheckConcern 检查是否有任意一个群添加过id的ctype订阅，如果添加过，返回 ErrAlreadyExists
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

// AddGroupConcern 在group内添加id的ctype订阅，多次添加同样的订阅会返回 ErrAlreadyExists
func (c *StateManager) AddGroupConcern(groupCode int64, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	err = c.RWCover(func() error {
		if c.CheckGroupConcern(groupCode, id, ctype) == ErrAlreadyExists {
			return ErrAlreadyExists
		}
		groupStateKey := c.GroupConcernStateKey(groupCode, id)
		newCtype, err = c.upsertConcernType(groupStateKey, ctype)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	if c.useEmit {
		allCtype, err := c.GetConcern(id)
		if err != nil {
			logger.WithField("id", id).Errorf("GetConcern error %v", err)
		} else {
			c.emitQueue.Add(localutils.NewEmitE(id, allCtype))
		}
	}
	return
}

// RemoveGroupConcern 在group内删除id的ctype订阅，并返回删除后当前id的综合ctype，删除不存在的订阅会返回 buntdb.ErrNotFound
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
	if err != nil {
		return
	}
	if c.useEmit {
		allCtype, err := c.GetConcern(id)
		if err != nil {
			logger.WithField("id", id).Errorf("GetConcern error %v", err)
		} else {
			if allCtype.Empty() {
				c.emitQueue.Delete(id)
			} else {
				c.emitQueue.Update(localutils.NewEmitE(id, allCtype))
			}
		}
	}
	return
}

// RemoveAllByGroupCode 删除一个group内所有订阅
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

// GetGroupConcern 返回一个id在群内的所有 concern_type.Type
func (c *StateManager) GetGroupConcern(groupCode int64, id interface{}) (result concern_type.Type, err error) {
	val, err := c.Get(c.GroupConcernStateKey(groupCode, id))
	if err != nil {
		return
	}
	result = concern_type.FromString(val)
	return
}

// GetConcern 查询一个id在所有group内的 concern_type.Type
func (c *StateManager) GetConcern(id interface{}) (result concern_type.Type, err error) {
	var ctypes []concern_type.Type
	_, _, ctypes, err = c.ListConcernState(func(groupCode int64, _id interface{}, p concern_type.Type) bool {
		return id == _id
	})
	result = concern_type.Empty.Add(ctypes...)
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
	} {
		c.CreatePatternIndex(pattern, nil)
	}
	var groupSet = make(map[int64]interface{})
	if len(groups) == 0 && miraiBot.Instance != nil {
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
		} {
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
	if c.useEmit {
		c.emitQueue.Stop()
		close(c.emitChan)
	}
	c.cancelCtx()
	c.freshWg.Wait()
	close(c.eventChan)
	c.wg.Wait()
}

func (c *StateManager) Start() error {
	if c.freshFunc == nil {
		panic(fmt.Sprintf("StateManager %v: freshFunc not set", c.name))
	}
	if c.notifyGeneratorFunc == nil {
		panic(fmt.Sprintf("StateManager %v: notifyGenerator not set", c.name))
	}
	if c.dispatchFunc == nil {
		c.Logger().Trace("use default DispatchFunc")
		c.UseDispatchFunc(c.DefaultDispatch())
	}
	c.FreshIndex()
	if runtime.NumCPU() >= 3 {
		for i := 0; i < 3; i++ {
			go c.Dispatch(&c.wg, c.eventChan, c.notifyChan)
		}
	} else {
		go c.Dispatch(&c.wg, c.eventChan, c.notifyChan)
	}
	if c.useEmit {
		c.emitQueue.Start()
		_, ids, ctypes, err := c.ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
			return true
		})
		if err != nil {
			return err
		}
		ids, ctypes, err = c.GroupTypeById(ids, ctypes)
		if err != nil {
			return err
		}
		for index := range ids {
			c.emitQueue.Add(localutils.NewEmitE(ids[index], ctypes[index]))
		}
	}
	go c.Fresh(&c.freshWg, c.eventChan)
	return nil
}

// EmitQueueFresher 如果使用的是EmitQueue，则可以使用这个helper来产生一个Fresher
func (c *StateManager) EmitQueueFresher(doFresh func(p concern_type.Type, id interface{}) ([]Event, error)) FreshFunc {
	return func(ctx context.Context, eventChan chan<- Event) {
		if !c.useEmit {
			panic(ErrEmitQueueNotInit)
		}
		for {
			select {
			case emitItem, received := <-c.emitChan:
				if !received {
					break
				}
				id := emitItem.Id
				if ok, _ := c.CheckFresh(id, true); !ok {
					logger.WithFields(logrus.Fields{
						"Id":     id,
						"Type":   emitItem.Type.String(),
						"Result": ok,
					}).Trace("fresh check failed")
					continue
				}
				logger.WithField("id", id).Trace("fresh")
				if events, err := doFresh(emitItem.Type, id); err == nil {
					for _, event := range events {
						c.eventChan <- event
					}
				} else {
					logger.WithFields(logrus.Fields{
						"Id":   id,
						"Type": emitItem.Type.String(),
						"Name": c.name,
					}).Errorf("doFresh error %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *StateManager) Fresh(wg *sync.WaitGroup, eventChan chan<- Event) {
	defer func() {
		if e := recover(); e != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("StateManager %v: Fresh panic recoved", c.name)
			go c.Fresh(wg, eventChan)
		}
	}()
	wg.Add(1)
	defer wg.Done()
	c.freshFunc(c.ctx, eventChan)
}

func (c *StateManager) Dispatch(wg *sync.WaitGroup, eventChan <-chan Event, notifyChan chan<- Notify) {
	defer func() {
		if e := recover(); e != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("StateManager %v: Dispatch panic recoved", c.name)
			go c.Dispatch(wg, eventChan, notifyChan)
		}
	}()
	wg.Add(1)
	defer wg.Done()
	c.dispatchFunc(eventChan, notifyChan)
}

func (c *StateManager) NotifyGenerator(groupCode int64, event Event) []Notify {
	return c.notifyGeneratorFunc(groupCode, event)
}

func (c *StateManager) DefaultDispatch() DispatchFunc {
	return func(eventChan <-chan Event, notifyChan chan<- Notify) {
		for event := range eventChan {
			log := event.Logger()
			groups, _, _, err := c.ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
				return event.GetUid() == id && p.ContainAll(event.Type())
			})
			if err != nil {
				log.Errorf("StateManager %v: ListConcernState error %v", c.name, err)
				continue
			}
			log.Debugf("new event - %v - %v notify for %v groups", event.Site(), event.Type().String(), len(groups))
			for _, groupCode := range groups {
				for _, n := range c.NotifyGenerator(groupCode, event) {
					notifyChan <- n
				}
			}
		}
	}
}

func (c *StateManager) UseFreshFunc(freshFunc FreshFunc) {
	c.freshFunc = freshFunc
}

func (c *StateManager) UseNotifyGeneratorFunc(notifyGeneratorFunc NotifyGeneratorFunc) {
	c.notifyGeneratorFunc = notifyGeneratorFunc
}

func (c *StateManager) UseDispatchFunc(dispatchFunc DispatchFunc) {
	c.dispatchFunc = dispatchFunc
}

var defaultInterval = time.Second * 5

// UseEmitQueue 启用EmitQueue
func (c *StateManager) UseEmitQueue() {
	c.useEmit = true
	var interval time.Duration
	if config.GlobalConfig != nil {
		interval = config.GlobalConfig.GetDuration("concern.emitInterval")
	}
	if interval == 0 {
		interval = defaultInterval
	}
	c.emitChan = make(chan *localutils.EmitE)
	c.emitQueue = localutils.NewEmitQueue(c.emitChan, interval)
}

func (c *StateManager) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Name": c.name,
	})
}

func NewStateManagerWithCustomKey(name string, keySet KeySet, notifyChan chan<- Notify) *StateManager {
	ctx, cancel := context.WithCancel(context.Background())
	sm := &StateManager{
		name:       name,
		notifyChan: notifyChan,
		eventChan:  make(chan Event, 64),
		KeySet:     keySet,
		ctx:        ctx,
		cancelCtx:  cancel,
	}
	return sm
}

func NewStateManagerWithStringID(name string, notifyChan chan<- Notify) *StateManager {
	return NewStateManagerWithCustomKey(name, NewPrefixKeySetWithStringID(name), notifyChan)
}

func NewStateManagerWithInt64ID(name string, notifyChan chan<- Notify) *StateManager {
	return NewStateManagerWithCustomKey(name, NewPrefixKeySetWithInt64ID(name), notifyChan)
}
