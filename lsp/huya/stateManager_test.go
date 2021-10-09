package huya

import (
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func initStateManager(t *testing.T) *StateManager {
	sm := NewStateManager()
	assert.NotNil(t, sm)
	assert.Nil(t, sm.Start())
	return sm
}

func TestNewStateManager(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)
}

func TestStateManager_GetLiveInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)

	expected := &LiveInfo{
		RoomId:   test.NAME1,
		Name:     test.NAME2,
		RoomName: test.NAME2,
	}
	assert.Nil(t, sm.AddLiveInfo(expected))
	actual, err := sm.GetLiveInfo(test.NAME1)
	assert.Nil(t, err)
	assert.EqualValues(t, expected, actual)
}
