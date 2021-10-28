package youtube

import (
	"github.com/Sora233/DDBOT/internal/test"
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

func TestStateManager_GetInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)

	expected := &Info{
		UserInfo: UserInfo{
			ChannelId:   test.NAME1,
			ChannelName: test.NAME2,
		},
	}
	assert.Nil(t, sm.AddInfo(expected))
	actual, err := sm.GetInfo(test.NAME1)
	assert.Nil(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestStateManager_GetVideo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)

	expected := &VideoInfo{
		UserInfo: UserInfo{
			ChannelId:   test.NAME1,
			ChannelName: test.NAME2,
		},
		VideoId:     test.BVID1,
		VideoType:   VideoType_Video,
		VideoStatus: VideoStatus_Upload,
	}

	assert.Nil(t, sm.AddVideo(expected))
	actual, err := sm.GetVideo(test.NAME1, test.BVID1)
	assert.Nil(t, err)
	assert.EqualValues(t, expected, actual)
}
