package douyu

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func initStateManager(t *testing.T) *StateManager {
	sm := NewStateManager(nil)
	assert.NotNil(t, sm)
	sm.FreshIndex(test.G1, test.G2)
	return sm
}
func TestNewStateManager(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)
	assert.NotNil(t, sm.GetGroupConcernConfig(test.G1, test.UID1))
}

func TestStateManager_GetLiveInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)

	var expected = &LiveInfo{
		Nickname: test.NAME1,
		RoomId:   test.UID1,
		RoomName: test.NAME2,
		RoomUrl:  "",
	}

	assert.Nil(t, sm.AddLiveInfo(expected))

	actual, err := sm.GetLiveInfo(test.UID1)
	assert.Nil(t, err)
	assert.EqualValues(t, expected, actual)

	assert.NotNil(t, sm.AddLiveInfo(nil))
}
