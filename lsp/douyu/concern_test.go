package douyu

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConcern(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testChan := make(chan concern.Notify)

	c := NewConcern(testChan)
	c.StateManager = initStateManager(t)

	go c.notifyLoop()

	const _testRoom = "9617408"
	testRoom, err := ParseUid(_testRoom)
	assert.Nil(t, err)

	_, err = c.Add(test.G1, testRoom, concern.DouyuLive)
	assert.Nil(t, err)

	liveInfo, err := c.FindOrLoadRoom(testRoom)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo)
	assert.Equal(t, testRoom, liveInfo.RoomId)
	assert.Equal(t, "斗鱼官方视频号", liveInfo.RoomName)

	liveInfoes, ctypes, err := c.ListWatching(test.G1, concern.DouyuLive)
	assert.Nil(t, err)
	assert.Len(t, liveInfoes, 1)
	assert.Len(t, ctypes, 1)
	assert.Equal(t, concern.DouyuLive, ctypes[0])

	liveInfo = liveInfoes[0]
	assert.Equal(t, testRoom, liveInfo.RoomId)
	assert.Equal(t, "斗鱼官方视频号", liveInfo.RoomName)

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
