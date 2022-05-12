package concern

import (
	"context"
	"errors"
	"fmt"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"go.uber.org/atomic"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("concern")
var ErrEmitQueueNotInit = errors.New("emit queue not init")
var ErrMaxGroupConcernExceed = errors.New("本群已达到订阅上限")

// NotifyGeneratorFunc 是 IStateManager.NotifyGenerator 函数的具体逻辑
// 它针对一组 groupCode 把 Event 转变成一组 Notify
//
// 使用 StateManager 时，在 StateManager.Start 之前，
// 必须使用 StateManager.UseNotifyGeneratorFunc 来指定一个 NotifyGeneratorFunc, 否则会发生 panic
type NotifyGeneratorFunc func(target mt.Target, event Event) []Notify

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
	GetConcernConfig(target mt.Target, id interface{}) (concernConfig IConfig)
	OperateConcernConfig(target mt.Target, id interface{}, cfg IConfig, f func(concernConfig IConfig) bool) error

	GetTargetConcern(target mt.Target, id interface{}) (result concern_type.Type, err error)
	GetConcern(id interface{}) (result concern_type.Type, err error)

	CheckAndSetAtAllMark(target mt.Target, id interface{}) (result bool)
	CheckTargetConcern(target mt.Target, id interface{}, ctype concern_type.Type) error
	CheckConcern(id interface{}, ctype concern_type.Type) error

	AddTargetConcern(target mt.Target, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error)
	RemoveTargetConcern(target mt.Target, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error)
	RemoveAllByTarget(target mt.Target) (keys []string, err error)

	ListConcernState(filter func(target mt.Target, id interface{}, p concern_type.Type) bool) (targets []mt.Target,
		ids []interface{}, idTypes []concern_type.Type, err error)
	GroupTypeById(ids []interface{}, types []concern_type.Type) ([]interface{}, []concern_type.Type, error)

	// NotifyGenerator 从 Event 产生多个 Notify
	NotifyGenerator(target mt.Target, event Event) []Notify
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
	logger              *logrus.Entry
	maxGroupConcern     int
	largeNotifyCount    atomic.Int32
}

func (c *StateManager) getConcernConfig(target mt.Target, id interface{}) (concernConfig *ConcernConfig) {
	val, err := c.Get(c.ConcernConfigKey(target, id), localdb.IgnoreNotFoundOpt())
	if err != nil {
		c.Logger().WithFields(localutils.TargetFields(target)).
			WithField("id", id).
			Errorf("GetConcernConfig error %v", err)
	}
	if len(val) > 0 {
		concernConfig, err = NewConcernConfigFromString(val)
		if err != nil {
			c.Logger().WithFields(localutils.TargetFields(target)).
				WithFields(logrus.Fields{"id": id, "val": val}).Errorf("NewConcernConfigFromString error %v", err)
		}
	}
	if concernConfig == nil {
		concernConfig = new(ConcernConfig)
	}
	if concernConfig.ConcernAtMap == nil {
		concernConfig.ConcernAtMap = make(map[mt.TargetType]*ConcernAtConfig)
	}
	if concernConfig.ConcernFilterMap == nil {
		concernConfig.ConcernFilterMap = make(map[mt.TargetType]*ConcernFilterConfig)
	}
	if concernConfig.ConcernNotifyMap == nil {
		concernConfig.ConcernNotifyMap = make(map[mt.TargetType]*ConcernNotifyConfig)
	}
	return
}

// GetConcernConfig 总是返回non-nil
func (c *StateManager) GetConcernConfig(target mt.Target, id interface{}) IConfig {
	return c.getConcernConfig(target, id)
}

// OperateConcernConfig 在一个rw事务中获取GroupConcernConfig并交给函数，如果返回true，就保存GroupConcernConfig，否则就回滚。
func (c *StateManager) OperateConcernConfig(target mt.Target, id interface{}, cfg IConfig, f func(concernConfig IConfig) bool) error {
	err := c.RWCover(func() error {
		if !f(cfg) {
			return localdb.ErrRollback
		}
		if err := cfg.Validate(); err != nil {
			return err
		}
		ccfg := c.getConcernConfig(target, id)
		ccfg.ConcernNotifyMap[target.GetTargetType()] = cfg.GetConcernNotify(target.GetTargetType())
		ccfg.ConcernAtMap[target.GetTargetType()] = cfg.GetConcernAt(target.GetTargetType())
		ccfg.ConcernFilterMap[target.GetTargetType()] = cfg.GetConcernFilter(target.GetTargetType())
		return c.SetJson(c.ConcernConfigKey(target, id), ccfg)
	})
	return err
}

// CheckAndSetAtAllMark 检查@全体标记是否过期，未设置过或已过期返回true，并重置标记，否则返回false。
// 因为@全体有次数限制，并且较为恼人，故设置标记，两次@全体之间必须有间隔。
func (c *StateManager) CheckAndSetAtAllMark(target mt.Target, id interface{}) (result bool) {
	err := c.Set(c.AtAllMarkKey(target, id), "",
		localdb.SetExpireOpt(time.Hour*2), localdb.SetNoOverWriteOpt())
	return err == nil
}

// CheckTargetConcern 检查target是否已经添加过id的ctype订阅，如果添加过，返回 ErrAlreadyExists
func (c *StateManager) CheckTargetConcern(target mt.Target, id interface{}, ctype concern_type.Type) error {
	state, _ := c.GetTargetConcern(target, id)
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

// AddTargetConcern 在group内添加id的ctype订阅，多次添加同样的订阅会返回 ErrAlreadyExists，如果超过订阅上限，则会返回 ErrMaxGroupConcernExceed。
// 订阅上限可以使用 SetMaxGroupConcern 设置。
func (c *StateManager) AddTargetConcern(target mt.Target, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	err = c.RWCover(func() error {
		var err error
		if c.CheckTargetConcern(target, id, ctype) == ErrAlreadyExists {
			return ErrAlreadyExists
		}

		if c.maxGroupConcern > 0 {
			_, ids, ctypes, err := c.ListConcernState(func(_target mt.Target, id interface{}, p concern_type.Type) bool {
				return _target.Equal(target)
			})
			if err != nil {
				return err
			}
			ids, ctypes, _ = c.GroupTypeById(ids, ctypes)
			if len(ids) >= c.maxGroupConcern {
				return ErrMaxGroupConcernExceed
			}
		}

		groupStateKey := c.ConcernStateKey(target, id)
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
			c.Logger().WithField("id", id).Errorf("GetTargetConcern error %v", err)
		} else {
			c.emitQueue.Add(localutils.NewEmitE(id, allCtype))
		}
	}
	return
}

// RemoveTargetConcern 在group内删除id的ctype订阅，并返回删除后当前id的在群内的ctype，删除不存在的订阅会返回 buntdb.ErrNotFound
func (c *StateManager) RemoveTargetConcern(target mt.Target, id interface{}, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	err = c.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		if c.CheckTargetConcern(target, id, ctype) != ErrAlreadyExists {
			return buntdb.ErrNotFound
		}
		groupStateKey := c.ConcernStateKey(target, id)
		newCtype, err = c.removeConcernType(groupStateKey, ctype)
		return err
	})
	if err != nil {
		return
	}
	if c.useEmit {
		allCtype, err := c.GetConcern(id)
		if err != nil {
			c.Logger().WithField("id", id).Errorf("GetTargetConcern error %v", err)
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

// RemoveAllByTarget 删除一个target内所有订阅
func (c *StateManager) RemoveAllByTarget(target mt.Target) (keys []string, err error) {
	var indexKey = []string{
		c.ConcernStateKey(),
		c.ConcernConfigKey(),
	}
	var prefixKey = []string{
		c.ConcernStateKey(target),
		c.ConcernConfigKey(target),
	}
	return localdb.RemoveByPrefixAndIndex(prefixKey, indexKey)
}

func (c *StateManager) RemoveAllById(_id interface{}) (err error) {
	return c.RWCoverTx(func(tx *buntdb.Tx) error {
		var removeKey []string
		var iterErr error
		iterErr = tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
			var id interface{}
			_, id, iterErr = c.ParseConcernStateKey(key)
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
		if c.useEmit {
			c.emitQueue.Delete(_id)
		}
		return nil
	})
}

// GetTargetConcern 返回一个id在target内的所有 concern_type.Type
func (c *StateManager) GetTargetConcern(target mt.Target, id interface{}) (result concern_type.Type, err error) {
	val, err := c.Get(c.ConcernStateKey(target, id))
	if err != nil {
		return
	}
	result = concern_type.FromString(val)
	return
}

// GetConcern 查询一个id在所有target_type内的 concern_type.Type
func (c *StateManager) GetConcern(id interface{}) (result concern_type.Type, err error) {
	var ctypes []concern_type.Type
	_, _, ctypes, err = c.ListConcernState(func(target mt.Target, _id interface{}, p concern_type.Type) bool {
		return id == _id
	})
	result = concern_type.Empty.Add(ctypes...)
	return
}

// ListConcernState 遍历所有订阅，并根据 filter 返回需要的订阅
func (c *StateManager) ListConcernState(filter func(target mt.Target, id interface{}, p concern_type.Type) bool) (targets []mt.Target, ids []interface{}, idTypes []concern_type.Type, err error) {
	err = c.RCoverTx(func(tx *buntdb.Tx) error {
		var iterErr error
		err := tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
			var target mt.Target
			var id interface{}
			target, id, iterErr = c.ParseConcernStateKey(key)
			if iterErr != nil {
				return false
			}
			ctype := concern_type.FromString(value)
			if ctype.Empty() {
				return true
			}
			if filter(target, id, ctype) == true {
				targets = append(targets, target)
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

// GroupTypeById 按id聚合ctype，通常是配合 ListConcernState 使用，把 ListConcernState 返回的订阅按id聚合
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

func (c *StateManager) checkFresh(id interface{}, setTTL bool) bool {
	freshKey := c.FreshKey(id)
	err := c.RWCover(func() error {
		if c.Exist(freshKey) {
			return ErrAlreadyExists
		}
		if setTTL {
			return c.Set(freshKey, "", localdb.SetExpireOpt(time.Minute))
		}
		return nil
	})
	return err == nil
}

// SetMaxGroupConcern 设置单个群订阅的数量上限，当设置为0或者负数表示无限制。
func (c *StateManager) SetMaxGroupConcern(maxGroupConcern int) {
	if maxGroupConcern < 0 {
		maxGroupConcern = 0
	}
	c.maxGroupConcern = maxGroupConcern
}

// FreshIndex 刷新 group 的 index，通常不需要用户主动调用
// 在单元测试中有时候需要主动刷新 index，否则遍历时会返回 buntdb.ErrNotFound
func (c *StateManager) FreshIndex(targets ...mt.Target) {
	for _, pattern := range []localdb.KeyPatternFunc{
		c.ConcernStateKey, c.ConcernConfigKey,
	} {
		c.CreatePatternIndex(pattern, nil)
	}
}

func (c *StateManager) upsertConcernType(key string, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	err = c.RWCover(func() error {
		val, err := c.Get(key, localdb.IgnoreNotFoundOpt())
		if err != nil {
			return err
		}
		newCtype = concern_type.FromString(val).Add(ctype)
		return c.Set(key, newCtype.String())
	})
	return
}

func (c *StateManager) removeConcernType(key string, ctype concern_type.Type) (newCtype concern_type.Type, err error) {
	err = c.RWCover(func() error {
		val, err := c.Get(key)
		if err != nil {
			return err
		}
		oldCtype := concern_type.FromString(val)
		if !oldCtype.ContainAll(ctype) {
			return buntdb.ErrNotFound
		}
		newCtype = oldCtype.Remove(ctype)
		if newCtype.Empty() {
			_, err = c.Delete(key)
		} else {
			err = c.Set(key, newCtype.String())
		}
		return err
	})
	return
}

// Stop 停止 StateManager，别忘记在 Concern.Stop 中停止
func (c *StateManager) Stop() {
	if c.useEmit {
		c.emitQueue.Stop()
		close(c.emitChan)
	}
	c.Logger().Trace("StateManager stopping")
	c.cancelCtx()
	c.Logger().Trace("wait freshWg")
	c.freshWg.Wait()
	close(c.eventChan)
	c.Logger().Trace("wait wg")
	c.wg.Wait()
}

// Start 启动 StateManager，别忘记在 Concern.Start 中启动
// 启动前需要指定 FreshFunc 与 NotifyGeneratorFunc，否则会panic
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
		_, ids, ctypes, err := c.ListConcernState(func(target mt.Target, id interface{}, p concern_type.Type) bool {
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
			panic("emit queue not init, did you call UseEmitQueue() ?")
		}
		for {
			select {
			case emitItem, received := <-c.emitChan:
				if !received {
					return
				}
				id := emitItem.Id
				if ok := c.checkFresh(id, true); !ok {
					c.Logger().WithFields(logrus.Fields{
						"Id":     id,
						"Type":   emitItem.Type.String(),
						"Result": ok,
					}).Trace("fresh check failed")
					continue
				}
				c.Logger().WithField("id", id).Trace("fresh")
				if events, err := doFresh(emitItem.Type, id); err == nil {
					for _, event := range events {
						c.eventChan <- event
					}
				} else {
					c.Logger().WithFields(logrus.Fields{
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
			c.Logger().WithField("stack", string(debug.Stack())).
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
			c.Logger().WithField("stack", string(debug.Stack())).
				Errorf("StateManager %v: Dispatch panic <%v> recoved", c.name, e)
			go c.Dispatch(wg, eventChan, notifyChan)
		}
	}()
	wg.Add(1)
	defer wg.Done()
	c.dispatchFunc(eventChan, notifyChan)
}

func (c *StateManager) NotifyGenerator(target mt.Target, event Event) []Notify {
	return c.notifyGeneratorFunc(target, event)
}

func (c *StateManager) filterNotify(inotify Notify) bool {
	if inotify == nil {
		return false
	}
	nLogger := inotify.Logger()
	concern, err := GetConcernBySiteAndType(inotify.Site(), inotify.Type())
	if err != nil {
		nLogger.Errorf("filterNotify: GetConcernBySiteAndType error %v", err)
		return true
	}
	concernConfig := concern.GetStateManager().GetConcernConfig(inotify.GetTarget(), inotify.GetUid())

	sendHookResult := concernConfig.ShouldSendHook(inotify)
	if !sendHookResult.Pass {
		nLogger.WithField("Reason", sendHookResult.Reason).Trace("notify filtered by hook ShouldSendHook")
		return false
	}

	newsFilterHook := concernConfig.FilterHook(inotify)
	if !newsFilterHook.Pass {
		nLogger.WithField("Reason", newsFilterHook.Reason).Trace("notify filtered by hook FilterHook")
		return false
	}
	return true
}

// DefaultDispatch 是 DispatchFunc 的默认实现。
// 它查询所有订阅过此 Event.GetUid 与 Event.Type 的群，并为每个群生成 Notify 发送给框架
func (c *StateManager) DefaultDispatch() DispatchFunc {
	return func(eventChan <-chan Event, notifyChan chan<- Notify) {
		for event := range eventChan {
			log := event.Logger()
			targets, _, _, err := c.ListConcernState(func(target mt.Target, id interface{}, p concern_type.Type) bool {
				return event.GetUid() == id && p.ContainAll(event.Type())
			})
			if err != nil {
				log.Errorf("StateManager %v: ListConcernState error %v", c.name, err)
				continue
			}
			var notifies []Notify
			var filteredGroups = make(map[string]interface{})
			for _, target := range targets {
				for _, n := range c.NotifyGenerator(target, event) {
					if c.filterNotify(n) {
						notifies = append(notifies, n)
						filteredGroups[n.GetTarget().Hash()] = true
					}
				}
			}
			if len(notifies) == 0 {
				continue
			}
			log.Infof("new event - %v %v - %v notify for %v groups", event.Site(), event.Type().String(), len(notifies), len(filteredGroups))
			largeNotifyLimit := cfg.GetLargeNotifyLimit()

			if len(notifies) >= largeNotifyLimit {
				log.Warnf("警告：当前事件将推送至%v条消息到%v个群（超过%v），为保证帐号稳定，将增加此事件的推送间隔，防止短时间内发送大量消息",
					len(notifies), len(filteredGroups), largeNotifyLimit)
				go func(notifies []Notify) {
					cnt := c.largeNotifyCount.Inc()
					ticker := time.NewTicker(time.Second*1 + time.Second*time.Duration(2*cnt))
					for _, n := range notifies {
						<-ticker.C
						notifyChan <- n
					}
					ticker.Stop()
					c.largeNotifyCount.Dec()
				}(notifies)
			} else {
				for _, n := range notifies {
					notifyChan <- n
				}
			}
		}
	}
}

// UseFreshFunc 指定 FreshFunc，启动前必须指定，否则会panic
func (c *StateManager) UseFreshFunc(freshFunc FreshFunc) {
	c.freshFunc = freshFunc
}

// UseNotifyGeneratorFunc 指定 NotifyGeneratorFunc，启动前必须指定，否则会panic
func (c *StateManager) UseNotifyGeneratorFunc(notifyGeneratorFunc NotifyGeneratorFunc) {
	c.notifyGeneratorFunc = notifyGeneratorFunc
}

// UseDispatchFunc 指定 DispatchFunc，如果启动时没有指定，则会使用默认实现 DefaultDispatch
func (c *StateManager) UseDispatchFunc(dispatchFunc DispatchFunc) {
	c.dispatchFunc = dispatchFunc
}

var defaultInterval = time.Second * 5

// UseEmitQueue 启用EmitQueue
func (c *StateManager) UseEmitQueue() {
	c.useEmit = true
	var interval = cfg.GetEmitInterval()
	if interval == 0 {
		interval = defaultInterval
	}
	c.emitChan = make(chan *localutils.EmitE)
	c.emitQueue = localutils.NewEmitQueue(c.emitChan, interval)
}

// EmitQueueEnabled 返回是否使用了EmitQueue
func (c *StateManager) EmitQueueEnabled() bool {
	return c.useEmit
}

func (c *StateManager) Logger() *logrus.Entry {
	return c.logger
}

// NewStateManagerWithCustomKey 使用自定义的 KeySet 创建 StateManager，
// 如果不关心 KeySet，推荐使用 NewStateManagerWithStringID 或者 NewStateManagerWithInt64ID
// name 可以简单地使用 Concern.Site
func NewStateManagerWithCustomKey(name string, keySet KeySet, notifyChan chan<- Notify) *StateManager {
	ctx, cancel := context.WithCancel(context.Background())
	sm := &StateManager{
		name:       name,
		notifyChan: notifyChan,
		eventChan:  make(chan Event, 4),
		KeySet:     keySet,
		ctx:        ctx,
		cancelCtx:  cancel,
		logger:     logger.WithFields(logrus.Fields{"Name": name}),
	}
	return sm
}

// NewStateManagerWithStringID 创建新的 StateManager，会使用 NewPrefixKeySetWithStringID 创建 KeySet
// ID的格式必须与 Concern.ParseId 返回的格式匹配
// name 可以简单地使用 Concern.Site
func NewStateManagerWithStringID(name string, notifyChan chan<- Notify) *StateManager {
	return NewStateManagerWithCustomKey(name, NewPrefixKeySetWithStringID(name), notifyChan)
}

// NewStateManagerWithInt64ID 创建新的 StateManager，会使用 NewPrefixKeySetWithInt64ID 创建 KeySet
// ID的格式必须与 Concern.ParseId 返回的格式匹配
// name 可以简单地使用 Concern.Site
func NewStateManagerWithInt64ID(name string, notifyChan chan<- Notify) *StateManager {
	return NewStateManagerWithCustomKey(name, NewPrefixKeySetWithInt64ID(name), notifyChan)
}
