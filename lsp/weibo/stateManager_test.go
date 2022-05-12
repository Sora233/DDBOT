package weibo

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func initStateManager(t *testing.T) *StateManager {
	sm := NewStateManager(nil)
	assert.NotNil(t, sm)
	sm.FreshIndex()
	return sm
}

func TestNewStateManager(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	initStateManager(t)
}

func TestStateManager_GetNewsInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm.AddUserInfo(nil))
	assert.NotNil(t, sm.AddNewsInfo(nil))

	userInfo := &UserInfo{
		Uid:  test.UID1,
		Name: test.NAME1,
	}
	newsInfo := &NewsInfo{
		UserInfo: userInfo,
	}
	assert.Nil(t, sm.AddUserInfo(userInfo))
	assert.Nil(t, sm.AddNewsInfo(newsInfo))

	actualUserInfo, err := sm.GetUserInfo(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, userInfo, actualUserInfo)

	actualNewsInfo, err := sm.GetNewsInfo(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, newsInfo, actualNewsInfo)
}

func TestStateManager_MarkMblogId(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	replaced, err := sm.MarkMblogId(test.BVID1)
	assert.Nil(t, err)
	assert.False(t, replaced)

	replaced, err = sm.MarkMblogId(test.BVID1)
	assert.Nil(t, err)
	assert.True(t, replaced)
}
