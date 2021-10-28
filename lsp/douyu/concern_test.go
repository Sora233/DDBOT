package douyu

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const _testRoom = "9617408"

func TestConcern(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testChan := make(chan concern.Notify)

	c := NewConcern(testChan)
	c.StateManager = initStateManager(t)
	defer c.Stop()

	go c.notifyLoop()

	testRoom, err := ParseUid(_testRoom)
	assert.Nil(t, err)

	_, err = c.Add(nil, test.G1, testRoom, Live)
	assert.Nil(t, err)

	liveInfo, err := c.FindOrLoadRoom(testRoom)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo)
	assert.Equal(t, testRoom, liveInfo.RoomId)
	assert.Equal(t, "斗鱼官方视频号", liveInfo.RoomName)

	identityInfos, ctypes, err := c.List(test.G1, Live)
	assert.Nil(t, err)
	assert.Len(t, identityInfos, 1)
	assert.Len(t, ctypes, 1)
	assert.Equal(t, Live, ctypes[0])

	info := identityInfos[0]
	assert.Equal(t, testRoom, info.GetUid())
	assert.Equal(t, "斗鱼官方视频号", info.GetName())

	liveInfo.ShowStatus = ShowStatus_Living
	liveInfo.VideoLoop = VideoLoopStatus_Off

	c.eventChan <- liveInfo

	select {
	case notify := <-testChan:
		assert.Equal(t, test.G1, notify.GetGroupCode())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}
}
