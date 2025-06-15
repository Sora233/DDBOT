package youtube

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
}

func TestStateManager_GetInfo(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := initStateManager(t)
	assert.NotNil(t, sm)

	_, err := sm.GetInfo(test.NAME1)
	assert.NotNil(t, err)

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
