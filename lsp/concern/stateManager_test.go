package concern

import (
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"testing"
)

type testKeySet struct{}

func (t *testKeySet) GroupConcernStateKey(keys ...interface{}) string {
	return localdb.NamedKey("test1", keys)
}

func (t *testKeySet) GroupConcernConfigKey(keys ...interface{}) string {
	return localdb.NamedKey("test2", keys)
}

func (t *testKeySet) FreshKey(keys ...interface{}) string {
	return localdb.NamedKey("test3", keys)
}

func (t *testKeySet) GroupAtAllMarkKey(keys ...interface{}) string {
	return localdb.NamedKey("test4", keys)
}

func (t *testKeySet) ParseGroupConcernStateKey(key string) (groupCode int64, id interface{}, err error) {
	return localdb.ParseConcernStateKeyWithInt64(key)
}

func newStateManager(t *testing.T) *StateManager {
	sm := NewStateManager(&testKeySet{}, false)
	assert.NotNil(t, sm)
	sm.FreshIndex(test.G1, test.G2)
	return sm
}

func TestNewStateManager(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)

	sm.Start()
	sm.EmitFreshCore("name", func(ctype Type, id interface{}) error {
		return nil
	})
}

func TestStateManager_GroupConcernConfig(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)

	c := sm.GetGroupConcernConfig(test.G1, test.UID1)
	assert.NotNil(t, c)

	assert.Nil(t, c.GetGroupConcernAt().AtSomeone)
	assert.EqualValues(t, c, new(GroupConcernConfig))

	err := sm.OperateGroupConcernConfig(test.G1, test.UID1, func(concernConfig IConfig) bool {
		concernConfig.GetGroupConcernNotify().TitleChangeNotify = test.BibiliLive
		concernConfig.GetGroupConcernAt().AtSomeone = []*AtSomeone{
			{
				Ctype:  test.DouyuLive,
				AtList: []int64{1, 2, 3},
			},
		}
		concernConfig.GetGroupConcernAt().AtAll = test.YoutubeLive
		return true
	})
	assert.Nil(t, err)

	c = sm.GetGroupConcernConfig(test.G1, test.UID1)
	assert.NotNil(t, c)
	assert.EqualValues(t, c.GetGroupConcernNotify().TitleChangeNotify, test.BibiliLive)
	assert.EqualValues(t, c.GetGroupConcernAt().AtAll, test.YoutubeLive)
	assert.EqualValues(t, c.GetGroupConcernAt().AtSomeone, []*AtSomeone{
		{
			Ctype:  test.DouyuLive,
			AtList: []int64{1, 2, 3},
		},
	})

	err = sm.OperateGroupConcernConfig(test.G1, test.UID1, func(concernConfig IConfig) bool {
		concernConfig.GetGroupConcernNotify().TitleChangeNotify = Empty
		return false
	})
	assert.EqualValues(t, localdb.ErrRollback, err)

	c = sm.GetGroupConcernConfig(test.G1, test.UID1)
	assert.NotNil(t, c)
	assert.EqualValues(t, c.GetGroupConcernNotify().TitleChangeNotify, test.BibiliLive)
}

func TestStateManager_CheckAndSetAtAllMark(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)

	assert.True(t, sm.CheckAndSetAtAllMark(test.G1, test.UID1))
	assert.False(t, sm.CheckAndSetAtAllMark(test.G1, test.UID1))
}

func TestStateManager_FreshCheck(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)

	result, err := sm.CheckFresh(test.UID1, false)
	assert.True(t, result)
	assert.Nil(t, err)
	result, err = sm.CheckFresh(test.UID1, true)
	assert.True(t, result)
	assert.Nil(t, err)
	result, err = sm.CheckFresh(test.UID1, true)
	assert.False(t, result)
	assert.Nil(t, err)

}

func TestStateManager_GroupConcern(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)

	assert.Nil(t, sm.CheckGroupConcern(test.G1, test.UID1, test.BibiliLive))

	_, err := sm.AddGroupConcern(test.G1, test.UID1, test.BibiliLive.Add(test.YoutubeLive))
	assert.Nil(t, err)
	_, err = sm.AddGroupConcern(test.G2, test.UID1, test.HuyaLive)
	assert.Nil(t, err)

	_, err = sm.AddGroupConcern(test.G1, test.UID2, test.DouyuLive)
	assert.Nil(t, err)
	_, err = sm.AddGroupConcern(test.G2, test.UID2, test.DouyuLive)
	assert.Nil(t, err)

	ctype, err := sm.GetConcern(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.BibiliLive.Add(test.YoutubeLive).Add(test.HuyaLive), ctype)

	// G1 UID1: blive|ylive , UID2: dlive
	// G2 UID1: hlive       , UID2  dlive

	// 检查UID在G1中有 blive和ylive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckGroupConcern(test.G1, test.UID1, test.BibiliLive))
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckGroupConcern(test.G1, test.UID1, test.YoutubeLive))

	// 检查UID在G1没有 hlive和dlive
	assert.EqualValues(t, nil, sm.CheckGroupConcern(test.G1, test.UID1, test.DouyuLive))
	assert.EqualValues(t, nil, sm.CheckGroupConcern(test.G1, test.UID1, test.HuyaLive))

	// 检查UID在G2中有hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckGroupConcern(test.G2, test.UID1, test.HuyaLive))

	// 检查UID2 在G1和G2中有dlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckGroupConcern(test.G1, test.UID2, test.DouyuLive))
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckGroupConcern(test.G2, test.UID2, test.DouyuLive))

	// 检查UID2 在G1中没有blive和ylive
	assert.EqualValues(t, nil, sm.CheckGroupConcern(test.G1, test.UID2, test.BibiliLive))
	assert.EqualValues(t, nil, sm.CheckGroupConcern(test.G1, test.UID2, test.YoutubeLive))

	// 添加已有的状态不会报错
	_, err = sm.AddGroupConcern(test.G1, test.UID1, test.BibiliLive)
	assert.Nil(t, err)

	// 检查UID在所有G中有blive和ylive和hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckConcern(test.UID1, test.BibiliLive.Add(test.YoutubeLive).Add(test.HuyaLive)))
	// 检查UID在所有G中有没有dlive
	assert.Nil(t, sm.CheckConcern(test.UID1, test.DouyuLive))

	// 删除UID在G1中的ylive
	_, err = sm.RemoveGroupConcern(test.G1, test.UID1, test.YoutubeLive)
	assert.Nil(t, err)

	ctype, err = sm.GetConcern(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.BibiliLive.Add(test.HuyaLive), ctype)

	// 检查UID在所有G中没有ylive
	assert.Nil(t, sm.CheckConcern(test.UID1, test.YoutubeLive))
	// 检查UID在所有G中有blive和hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckConcern(test.UID1, test.BibiliLive.Add(test.HuyaLive)))
	// 检查UID在G1中有blive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckGroupConcern(test.G1, test.UID1, test.BibiliLive))
	// 检查UID在G2中有hlive
	assert.EqualValues(t, ErrAlreadyExists, sm.CheckGroupConcern(test.G2, test.UID1, test.HuyaLive))

	// 列出所有有hlive的记录，应该只有UID G2
	groups, ids, ctypes, err := sm.List(func(groupCode int64, id interface{}, p Type) bool {
		return p.ContainAny(test.HuyaLive)
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, 1, len(ids))
	assert.Equal(t, 1, len(ctypes))
	assert.Equal(t, test.G2, groups[0])
	assert.Equal(t, test.UID1, ids[0])
	assert.Equal(t, test.HuyaLive, ctypes[0])

	ctype, err = sm.GetGroupConcern(test.G2, test.UID2)
	assert.Nil(t, err)
	assert.EqualValues(t, test.DouyuLive, ctype)

	// G1中有 UID1:blive UID2:dlive
	ids, ctypes, err = sm.ListByGroup(test.G1, func(id interface{}, p Type) bool {
		return true
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

	ids, err = sm.ListIds()
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(ids))
	assert.Contains(t, ids, test.UID1)
	assert.Contains(t, ids, test.UID2)

	err = sm.RemoveAllById(test.UID1)
	assert.Nil(t, err)
	ctype, err = sm.GetConcern(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, ctype)

	err = sm.RemoveAllByGroupCode(test.G2)
	assert.Nil(t, err)
	ctype, err = sm.GetGroupConcern(test.G1, test.UID2)
	assert.Nil(t, err)
	assert.EqualValues(t, test.DouyuLive, ctype)
	ctype, err = sm.GetGroupConcern(test.G2, test.UID2)
	assert.EqualValues(t, buntdb.ErrNotFound, err)
}
