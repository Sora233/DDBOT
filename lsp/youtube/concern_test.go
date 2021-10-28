package youtube

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
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
	defer c.Stop()

	go c.notifyLoop()

	_, err := c.StateManager.AddGroupConcern(test.G1, test.NAME1, Live)
	assert.Nil(t, err)

	c.eventChan <- &VideoInfo{
		UserInfo: UserInfo{
			ChannelId:   test.NAME1,
			ChannelName: test.NAME1,
		},
		VideoType:         VideoType_Live,
		VideoStatus:       VideoStatus_Living,
		LiveStatusChanged: true,
	}

	select {
	case notify := <-testChan:
		assert.Equal(t, test.G1, notify.GetGroupCode())
		assert.Equal(t, test.NAME1, notify.GetUid())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}
}
