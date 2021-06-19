package bilibili

import (
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"testing"
)

func initStateManager(t *testing.T) *StateManager {
	sm := NewStateManager()
	assert.NotNil(t, sm)
	sm.FreshIndex(test.G1, test.G2)
	assert.Nil(t, sm.Start())
	return sm
}

func TestNewStateManager(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)
}

func TestStateManager_GetUserInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initStateManager(t)
	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	assert.NotNil(t, origUserInfo)
	err := c.AddUserInfo(origUserInfo)
	assert.Nil(t, err)

	userInfo, err := c.GetUserInfo(test.UID1)
	assert.EqualValues(t, origUserInfo, userInfo)

	assert.NotNil(t, c.AddUserInfo(nil))
}

func TestStateManager_GetLiveInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initStateManager(t)

	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	origLiveInfo := NewLiveInfo(origUserInfo, "", "", LiveStatus_Living)
	assert.NotNil(t, origLiveInfo)

	err := c.AddLiveInfo(origLiveInfo)
	assert.Nil(t, err)

	userInfo, err := c.GetUserInfo(test.UID1)
	assert.Nil(t, err)
	assert.NotNil(t, userInfo)
	assert.EqualValues(t, origUserInfo, userInfo)

	liveInfo, err := c.GetLiveInfo(test.UID1)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo)
	assert.EqualValues(t, origLiveInfo, liveInfo)

	liveInfo, err = c.GetLiveInfo(test.UID2)
	assert.Equal(t, buntdb.ErrNotFound, err)
	assert.Nil(t, liveInfo)

	assert.NotNil(t, c.AddLiveInfo(nil))
}

func TestStateManager_GetNewsInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initStateManager(t)

	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	origNewsInfo := NewNewsInfo(origUserInfo, test.DynamicID1, test.TIMESTAMP1)

	err := c.AddNewsInfo(origNewsInfo)
	assert.Nil(t, err)

	userInfo, err := c.GetUserInfo(test.UID1)
	assert.Nil(t, err)
	assert.NotNil(t, userInfo)
	assert.EqualValues(t, origUserInfo, userInfo)

	newsInfo, err := c.GetNewsInfo(test.UID1)
	assert.Nil(t, err)
	assert.NotNil(t, newsInfo)
	assert.EqualValues(t, newsInfo, origNewsInfo)

	newsInfo, err = c.GetNewsInfo(test.UID2)
	assert.Equal(t, buntdb.ErrNotFound, err)
	assert.Nil(t, newsInfo)

	err = c.DeleteNewsInfo(origNewsInfo)
	assert.Nil(t, err)

	newsInfo, err = c.GetNewsInfo(test.UID1)
	assert.Equal(t, buntdb.ErrNotFound, err)
	assert.Nil(t, newsInfo)

	assert.NotNil(t, c.AddNewsInfo(nil))
	assert.NotNil(t, c.DeleteNewsInfo(nil))
}

func TestStateManager_CheckDynamicId(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initStateManager(t)

	assert.True(t, c.CheckDynamicId(test.DynamicID1))

	replaced, err := c.MarkDynamicId(test.DynamicID1)
	assert.Nil(t, err)
	assert.False(t, replaced)

	assert.False(t, c.CheckDynamicId(test.DynamicID1))

	replaced, err = c.MarkDynamicId(test.DynamicID1)
	assert.Nil(t, err)
	assert.True(t, replaced)
}

func TestStateManager_IncNotLiveCount(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initStateManager(t)

	assert.Equal(t, 1, c.IncNotLiveCount(test.UID1))
	assert.Equal(t, 2, c.IncNotLiveCount(test.UID1))
	assert.Equal(t, 3, c.IncNotLiveCount(test.UID1))
	assert.Equal(t, 4, c.IncNotLiveCount(test.UID1))
	assert.Equal(t, 5, c.IncNotLiveCount(test.UID1))
	assert.Equal(t, 6, c.IncNotLiveCount(test.UID1))

	assert.Nil(t, c.ClearNotLiveCount(test.UID1))
	assert.Equal(t, 1, c.IncNotLiveCount(test.UID1))
	assert.Equal(t, 2, c.IncNotLiveCount(test.UID1))
}

func TestStateManager_SetUidFirstTimestampIfNotExist(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initStateManager(t)

	_, err := c.GetUidFirstTimestamp(test.UID2)
	assert.Equal(t, buntdb.ErrNotFound, err)

	assert.Nil(t, c.SetUidFirstTimestampIfNotExist(test.UID1, test.TIMESTAMP1))

	ts1, err := c.GetUidFirstTimestamp(test.UID1)
	assert.Nil(t, err)
	assert.Equal(t, test.TIMESTAMP1, ts1)

	assert.Nil(t, c.SetUidFirstTimestampIfNotExist(test.UID1, test.TIMESTAMP2))
	ts1, err = c.GetUidFirstTimestamp(test.UID1)
	assert.Nil(t, err)
	assert.Equal(t, test.TIMESTAMP1, ts1)

	assert.Nil(t, c.UnsetUidFirstTimestamp(test.UID1))

	ts1, err = c.GetUidFirstTimestamp(test.UID1)
	assert.Equal(t, buntdb.ErrNotFound, err)
}
