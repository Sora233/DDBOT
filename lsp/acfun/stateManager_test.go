package acfun

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"

	"github.com/Sora233/DDBOT/v2/internal/test"
	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	"github.com/Sora233/DDBOT/v2/lsp/concern"
)

func initStateManager(t *testing.T, notifyChan chan<- concern.Notify) *StateManager {
	sm := NewStateManager(notifyChan)
	assert.NotNil(t, sm)
	sm.FreshIndex(test.G1, test.G2)
	return sm
}

func TestStateManager(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testNotifyChan := make(chan concern.Notify)

	sm := initStateManager(t, testNotifyChan)
	assert.NotNil(t, sm)
	defer sm.Stop()

	assert.NotNil(t, sm.GetGroupConcernConfig(test.G1, test.UID1))

	userInfo := UserInfo{
		Uid:  test.UID1,
		Name: test.NAME1,
	}

	_, err := sm.GetUserInfo(test.UID1)
	assert.NotNil(t, err)

	err = sm.AddUserInfo(&userInfo)
	assert.Nil(t, err)

	result, err := sm.GetUserInfo(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, &userInfo, result)

	_, err = sm.GetLiveInfo(test.UID1)
	assert.NotNil(t, err)

	err = sm.AddLiveInfo(&LiveInfo{
		UserInfo: userInfo,
	})
	assert.Nil(t, err)

	liveInfo, err := sm.GetLiveInfo(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.UID1, liveInfo.Uid)

	err = sm.DeleteLiveInfo(test.UID1)
	assert.Nil(t, err)

	_, err = sm.GetLiveInfo(test.UID1)
	assert.NotNil(t, err)

	err = localdb.RWCoverTx(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(sm.NotLiveKey(test.UID1), "wrong", nil)
		return err
	})
	assert.Nil(t, err)
	assert.Zero(t, sm.IncNotLiveCount(test.UID1))
	err = sm.ClearNotLiveCount(test.UID1)
	assert.Nil(t, err)

	for i := 1; i <= 10; i++ {
		assert.EqualValues(t, i, sm.IncNotLiveCount(test.UID1))
	}
	err = sm.ClearNotLiveCount(test.UID1)
	assert.Nil(t, err)
	for i := 1; i <= 10; i++ {
		assert.EqualValues(t, i, sm.IncNotLiveCount(test.UID1))
	}

	err = sm.SetUidFirstTimestampIfNotExist(test.UID1, test.TIMESTAMP1)
	assert.Nil(t, err)
	err = sm.SetUidFirstTimestampIfNotExist(test.UID1, test.TIMESTAMP2)
	assert.True(t, localdb.IsRollback(err))

	ts, err := sm.GetUidFirstTimestamp(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.TIMESTAMP1, ts)

	_, err = sm.GetUidFirstTimestamp(test.UID2)
	assert.NotNil(t, err)
}
