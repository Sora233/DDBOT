package huya

import (
	"context"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const testRoom = "s"

func TestConcern(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)

	c := NewConcern(testNotifyChan)

	assert.NotNil(t, c.GetStateManager())

	_, err := c.ParseId(testRoom)
	assert.Nil(t, err)

	c.StateManager.UseNotifyGeneratorFunc(c.notifyGenerator())
	c.StateManager.UseFreshFunc(func(ctx context.Context, eventChan chan<- concern.Event) {
		for {
			select {
			case e := <-testEventChan:
				if e != nil {
					eventChan <- e
				}
			case <-ctx.Done():
				return
			}
		}
	})

	assert.Nil(t, c.StateManager.Start())
	defer c.Stop()
	defer close(testEventChan)

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

	liveInfo2.liveStatusChanged = true
	liveInfo2.IsLiving = true

	testEventChan <- liveInfo2

	select {
	case notify := <-testNotifyChan:
		assert.Equal(t, test.G1, notify.GetGroupCode())
		assert.Equal(t, testRoom, notify.GetUid())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}

	_, err = c.Remove(nil, test.G1, testRoom, Live)
	assert.Nil(t, err)
}
