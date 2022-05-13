package concern

import (
	"context"
	"errors"
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"go.uber.org/atomic"
	"testing"
	"time"
)

const testSite = "test"

const (
	testType concern_type.Type = "test"
)

type testKeySet struct{}

func (t *testKeySet) ConcernStateKey(keys ...interface{}) string {
	return localdb.NamedKey("test1", keys)
}

func (t *testKeySet) ConcernConfigKey(keys ...interface{}) string {
	return localdb.NamedKey("test2", keys)
}

func (t *testKeySet) FreshKey(keys ...interface{}) string {
	return localdb.NamedKey("test3", keys)
}

func (t *testKeySet) AtAllMarkKey(keys ...interface{}) string {
	return localdb.NamedKey("test4", keys)
}

func (t *testKeySet) ParseConcernStateKey(key string) (target mt.Target, id interface{}, err error) {
	return ParseConcernStateKeyWithInt64(key)
}

type testEvent struct {
	id     int64
	target mt.Target
}

func (t *testEvent) GetTarget() mt.Target {
	return t.target
}

func (t *testEvent) ToMessage() *mmsg.MSG {
	return mmsg.NewTextf("test - id %v", t.id)
}

func (t *testEvent) Site() string {
	return testSite
}

func (t *testEvent) Type() concern_type.Type {
	return testType
}

func (t *testEvent) GetUid() interface{} {
	return t.id
}

func (t *testEvent) Logger() *logrus.Entry {
	return logrus.WithField("id", t.id)
}

func newStateManager(t *testing.T) *StateManager {
	sm := NewStateManagerWithCustomKey("test", &testKeySet{}, nil)
	assert.NotNil(t, sm)
	sm.FreshIndex(mt.NewGroupTarget(test.G1), mt.NewGroupTarget(test.G2))
	return sm
}

func TestNewStateManagerWithStringID(t *testing.T) {
	assert.NotNil(t, NewStateManagerWithStringID("test-string", nil))
}

func TestNewStateManagerWithInt64ID(t *testing.T) {
	assert.NotNil(t, NewStateManagerWithInt64ID("test-string", nil))
}

func TestNewStateManager(t *testing.T) {
	_defaultInterval := defaultInterval
	defaultInterval = time.Second
	defer func() {
		defaultInterval = _defaultInterval
	}()
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	sm.UseDispatchFunc(sm.DefaultDispatch())
	assert.Panics(t, func() {
		sm.Start()
	})
	emitHook := make(chan Event, 16)
	var freshError atomic.Bool
	freshError.Store(false)
	sm.UseFreshFunc(sm.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]Event, error) {
		if freshError.Load() {
			return nil, errors.New("fresh error")
		}
		e := &testEvent{id: id.(int64)}
		emitHook <- e
		return []Event{e}, nil
	}))
	assert.Panics(t, func() {
		sm.Start()
	})
	sm.UseNotifyGeneratorFunc(func(target mt.Target, event Event) []Notify {
		return nil
	})
	sm.UseEmitQueue()

	_, err := sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, "test")
	assert.Nil(t, err)
	sm.Start()
	defer sm.Stop()

	select {
	case e := <-emitHook:
		assert.EqualValues(t, test.UID1, e.GetUid())
	case <-time.After(time.Second * 2):
		assert.Fail(t, "no item received")
	}

	select {
	case <-emitHook:
		assert.Fail(t, "should no item received")
	case <-time.After(time.Second * 2):
	}

	err = localdb.RWCoverTx(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(sm.FreshKey(test.UID1))
		return err
	})
	assert.Nil(t, err)

	freshError.Store(true)

	select {
	case <-emitHook:
		assert.Fail(t, "should no item received")
	case <-time.After(time.Second * 2):
	}
}

func TestNewStateManager2(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	var testCount atomic.Int32
	sm.UseFreshFunc(func(ctx context.Context, eventChan chan<- Event) {
		if testCount.CAS(0, 1) {
			panic("error")
		}
	})
	sm.UseDispatchFunc(func(event <-chan Event, notify chan<- Notify) {
		if testCount.CAS(1, 2) {
			panic("error")
		}
	})
	sm.UseNotifyGeneratorFunc(func(target mt.Target, event Event) []Notify {
		return nil
	})
	assert.Nil(t, sm.Start())
	time.Sleep(time.Second)
}

func TestStateManagerNotify(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	var err error
	sm := newStateManager(t)

	testEventChan := make(chan Event, 16)
	testNotifyChan := make(chan Notify, 16)
	sm.notifyChan = testNotifyChan
	sm.UseNotifyGeneratorFunc(func(target mt.Target, event Event) []Notify {
		event.(*testEvent).target = target
		return []Notify{
			event.(*testEvent),
		}
	})
	sm.UseFreshFunc(func(ctx context.Context, eventChan chan<- Event) {
		for {
			select {
			case e := <-testEventChan:
				if e != nil {
					eventChan <- e
				}
			case <-ctx.Done():
				return
			}
		}
	})
	sm.Start()

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, testType)
	assert.Nil(t, err)
	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G2), test.UID1, testType)
	assert.Nil(t, err)
	testEventChan <- &testEvent{
		id: test.UID2,
	}

	select {
	case <-testNotifyChan:
		assert.Fail(t, "should no notify received")
	case <-time.After(time.Second):
	}

	testEventChan <- &testEvent{
		id: test.UID1,
	}

	for i := 0; i < 2; i++ {
		select {
		case notify := <-testNotifyChan:
			assert.NotNil(t, notify)
			assert.EqualValues(t, test.UID1, notify.GetUid())
			assert.True(t, notify.GetTarget().Equal(mt.NewGroupTarget(test.G1)) || notify.GetTarget().Equal(mt.NewGroupTarget(test.G2)))
		case <-time.After(time.Second):
			assert.Fail(t, "no item received")
		}
	}

}

func TestStateManager_GroupConcernConfig(t *testing.T) {
	sm := newStateManager(t)

	c := sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	assert.NotNil(t, c)

	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm = newStateManager(t)

	c = sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	assert.NotNil(t, c)

	assert.Nil(t, c.GetConcernAt().AtSomeone)

	cfg := sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	err := sm.OperateConcernConfig(mt.NewGroupTarget(test.G1), test.UID1, cfg, func(concernConfig IConfig) bool {
		concernConfig.GetConcernNotify().TitleChangeNotify = test.BibiliLive
		concernConfig.GetConcernAt().AtSomeone = []*AtSomeone{
			{
				Ctype:  test.DouyuLive,
				AtList: []int64{1, 2, 3},
			},
		}
		concernConfig.GetConcernAt().AtAll = test.YoutubeLive
		return true
	})
	assert.Nil(t, err)

	c = sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	assert.NotNil(t, c)
	assert.NotNil(t, c.GetConcernFilter())
	assert.EqualValues(t, c.GetConcernNotify().TitleChangeNotify, test.BibiliLive)
	assert.EqualValues(t, c.GetConcernAt().AtAll, test.YoutubeLive)
	assert.EqualValues(t, c.GetConcernAt().AtSomeone, []*AtSomeone{
		{
			Ctype:  test.DouyuLive,
			AtList: []int64{1, 2, 3},
		},
	})

	cfg = sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	err = sm.OperateConcernConfig(mt.NewGroupTarget(test.G1), test.UID1, cfg, func(concernConfig IConfig) bool {
		concernConfig.GetConcernNotify().TitleChangeNotify = concern_type.Empty
		return false
	})
	assert.EqualValues(t, localdb.ErrRollback, err)

	c = sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	assert.NotNil(t, c)
	assert.EqualValues(t, c.GetConcernNotify().TitleChangeNotify, test.BibiliLive)

	cfg = sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	err = sm.OperateConcernConfig(mt.NewGroupTarget(test.G1), test.UID1, cfg, func(concernConfig IConfig) bool {
		concernConfig.GetConcernFilter().Type = FilterTypeType
		concernConfig.GetConcernFilter().Config = (&GroupConcernFilterConfigByType{Type: []string{"q", "w", "e"}}).ToString()
		return true
	})
	assert.EqualValues(t, ErrConfigNotSupported, err)
}

func TestStateManager_CheckAndSetAtAllMark(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)

	assert.True(t, sm.CheckAndSetAtAllMark(mt.NewGroupTarget(test.G1), test.UID1))
	assert.False(t, sm.CheckAndSetAtAllMark(mt.NewGroupTarget(test.G1), test.UID1))
}

func TestStateManager_FreshCheck(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)

	result := sm.checkFresh(test.UID1, false)
	assert.True(t, result)
	result = sm.checkFresh(test.UID1, false)
	assert.True(t, result)
	result = sm.checkFresh(test.UID1, true)
	assert.True(t, result)
	result = sm.checkFresh(test.UID1, true)
	assert.False(t, result)
}

func TestStateManager_GroupConcern(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	var err error
	sm := newStateManager(t)
	sm.UseEmitQueue()

	assert.Nil(t, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive))

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID2, test.HuyaLive)
	assert.Nil(t, err)
	_, err = sm.RemoveTargetConcern(mt.NewGroupTarget(test.G1), test.UID2, test.HuyaLive)
	assert.Nil(t, err)

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive.Add(test.YoutubeLive))
	assert.Nil(t, err)

	_, err = sm.RemoveTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive)
	assert.Nil(t, err)
	_, err = sm.RemoveTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive)
	assert.EqualValues(t, buntdb.ErrNotFound, err)
	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive.Add(test.YoutubeLive))
	assert.Nil(t, err)

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G2), test.UID1, test.HuyaLive)
	assert.Nil(t, err)
	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G2), test.UID1, test.HuyaLive)
	assert.EqualValues(t, ErrAlreadyExists, err)

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID2, test.DouyuLive)
	assert.Nil(t, err)
	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G2), test.UID2, test.DouyuLive)
	assert.Nil(t, err)

	ctype, err := sm.GetConcern(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.BibiliLive.Add(test.YoutubeLive, test.HuyaLive), ctype)

	// G1 UID1: blive|ylive , UID2: dlive
	// G2 UID1: hlive       , UID2  dlive

	// 检查UID在G1中有 blive和ylive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive))
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.YoutubeLive))

	// 检查UID在G1没有 hlive和dlive
	assert.EqualValues(t, nil, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.DouyuLive))
	assert.EqualValues(t, nil, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.HuyaLive))

	// 检查UID在G2中有hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckTargetConcern(mt.NewGroupTarget(test.G2), test.UID1, test.HuyaLive))

	// 检查UID2 在G1和G2中有dlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID2, test.DouyuLive))
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckTargetConcern(mt.NewGroupTarget(test.G2), test.UID2, test.DouyuLive))

	// 检查UID2 在G1中没有blive和ylive
	assert.EqualValues(t, nil, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID2, test.BibiliLive))
	assert.EqualValues(t, nil, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID2, test.YoutubeLive))

	// 添加已有的状态会报错
	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive)
	assert.EqualValues(t, ErrAlreadyExists, err)

	// 检查UID在所有G中有blive和ylive和hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckConcern(test.UID1, test.BibiliLive.Add(test.YoutubeLive).Add(test.HuyaLive)))
	// 检查UID在所有G中有没有dlive
	assert.Nil(t, sm.CheckConcern(test.UID1, test.DouyuLive))

	// 删除UID在G1中的ylive
	_, err = sm.RemoveTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.YoutubeLive)
	assert.Nil(t, err)

	ctype, err = sm.GetConcern(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.BibiliLive.Add(test.HuyaLive), ctype)

	// 检查UID在所有G中没有ylive
	assert.Nil(t, sm.CheckConcern(test.UID1, test.YoutubeLive))
	// 检查UID在所有G中有blive和hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckConcern(test.UID1, test.BibiliLive.Add(test.HuyaLive)))
	// 检查UID在G1中有blive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive))
	// 检查UID在G2中有hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckTargetConcern(mt.NewGroupTarget(test.G2), test.UID1, test.HuyaLive))

	// 列出所有有hlive的记录，应该只有UID G2
	groups, ids, ctypes, err := sm.ListConcernState(func(target mt.Target, id interface{}, p concern_type.Type) bool {
		return p.ContainAny(test.HuyaLive)
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, 1, len(ids))
	assert.Equal(t, 1, len(ctypes))
	assert.True(t, groups[0].Equal(mt.NewGroupTarget(test.G2)))
	assert.Equal(t, test.UID1, ids[0])
	assert.Equal(t, test.HuyaLive, ctypes[0])

	ctype, err = sm.GetTargetConcern(mt.NewGroupTarget(test.G2), test.UID2)
	assert.Nil(t, err)
	assert.EqualValues(t, test.DouyuLive, ctype)

	// G1中有 UID1:blive UID2:dlive
	_, ids, ctypes, err = sm.ListConcernState(func(target mt.Target, id interface{}, p concern_type.Type) bool {
		return target.Equal(mt.NewGroupTarget(test.G1))
	})
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(ids))
	assert.EqualValues(t, len(ids), len(ctypes))
	for index := range ids {
		if ids[index] == test.UID1 {
			assert.EqualValues(t, test.BibiliLive, ctypes[index])
		} else {
			assert.EqualValues(t, test.DouyuLive, ctypes[index])
		}
	}

	_, _, err = sm.GroupTypeById([]interface{}{test.UID1}, nil)
	assert.EqualValues(t, ErrLengthMismatch, err)

	ids, ctypes, err = sm.GroupTypeById(ids, ctypes)
	assert.Nil(t, err)

	ids, err = listIds(sm)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(ids))
	assert.Contains(t, ids, test.UID1)
	assert.Contains(t, ids, test.UID2)

	err = sm.RemoveAllById(test.UID1)
	assert.Nil(t, err)
	ctype, err = sm.GetConcern(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, concern_type.Empty, ctype)

	_, err = sm.RemoveAllByTarget(mt.NewGroupTarget(test.G2))
	assert.Nil(t, err)
	ctype, err = sm.GetTargetConcern(mt.NewGroupTarget(test.G1), test.UID2)
	assert.Nil(t, err)
	assert.EqualValues(t, test.DouyuLive, ctype)
	ctype, err = sm.GetTargetConcern(mt.NewGroupTarget(test.G2), test.UID2)
	assert.EqualValues(t, buntdb.ErrNotFound, err)
}

func TestStateManager_GroupConcern2(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	var err error
	sm := newStateManager(t)
	sm.UseEmitQueue()
	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BilibiliNews)
	assert.Nil(t, err)

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.BibiliLive)
	assert.Nil(t, err)

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.DouyuLive)
	assert.Nil(t, err)

	_, err = sm.AddTargetConcern(mt.NewGroupTarget(test.G1), test.UID1, test.HuyaLive)
	assert.Nil(t, err)

	ctype, err := sm.GetTargetConcern(mt.NewGroupTarget(test.G1), test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.BilibiliNews.Add(test.BibiliLive, test.DouyuLive, test.HuyaLive), ctype)
}

func listIds(sm *StateManager) ([]interface{}, error) {
	var m = make(map[interface{}]interface{})
	_, _, _, err := sm.ListConcernState(func(target mt.Target, id interface{}, p concern_type.Type) bool {
		m[id] = struct{}{}
		return true
	})
	if err != nil {
		return nil, err
	}
	var ids []interface{}
	for k := range m {
		ids = append(ids, k)
	}
	return ids, nil
}
