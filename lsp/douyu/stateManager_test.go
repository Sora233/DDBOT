package douyu

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
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

	sm := initStateManager(t)
	assert.NotNil(t, sm)
	assert.NotNil(t, sm.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1))
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
