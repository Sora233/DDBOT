package huya

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const testRoom = "s"

func TestConcern(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testChan := make(chan concern.Notify)

	c := NewConcern(testChan)
	c.StateManager = initStateManager(t)
	defer c.Stop()

	go c.notifyLoop()

	_, err := c.Add(test.G1, testRoom, concern.HuyaLive)
	assert.Nil(t, err)

	liveInfoes, _, err := c.ListWatching(test.G1, concern.HuyaLive)
	assert.Nil(t, err)
	assert.Len(t, liveInfoes, 1)
	liveInfo := liveInfoes[0]

	liveInfo2, err := c.FindOrLoadRoom(testRoom)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo2)
	assert.EqualValues(t, liveInfo, liveInfo2)

	liveInfo.LiveStatusChanged = true
	liveInfo.Living = true

	c.eventChan <- liveInfo

	select {
	case notify := <-testChan:
		assert.Equal(t, test.G1, notify.GetGroupCode())
		assert.Equal(t, testRoom, notify.GetUid())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}
}
