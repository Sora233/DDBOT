package huya

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
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

	assert.NotNil(t, c.GetStateManager())

	_, err := c.ParseId(testRoom)
	assert.Nil(t, err)

	go c.notifyLoop()

	_, err = c.Add(nil, test.G1, testRoom, Live)
	assert.Nil(t, err)

	identityInfo, err := c.Get(testRoom)
	assert.Nil(t, err)
	assert.EqualValues(t, testRoom, identityInfo.GetUid())

	identityInfos, _, err := c.List(test.G1, Live)
	assert.Nil(t, err)
	assert.Len(t, identityInfos, 1)
	info := identityInfos[0]

	liveInfo2, err := c.FindOrLoadRoom(testRoom)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo2)
	assert.EqualValues(t, info.GetUid(), liveInfo2.RoomId)
	assert.EqualValues(t, info.GetName(), liveInfo2.GetName())

	liveInfo2.LiveStatusChanged = true
	liveInfo2.Living = true

	c.eventChan <- liveInfo2

	select {
	case notify := <-testChan:
		assert.Equal(t, test.G1, notify.GetGroupCode())
		assert.Equal(t, testRoom, notify.GetUid())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}

	_, err = c.Remove(nil, test.G1, testRoom, Live)
	assert.Nil(t, err)
}
