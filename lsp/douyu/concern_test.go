package douyu

import (
	"context"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const testRoomStr = "9617408"

func TestConcern(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)

	c := NewConcern(testNotifyChan)

	assert.NotNil(t, c.GetStateManager())

	_testRoom, err := c.ParseId(testRoomStr)
	assert.Nil(t, err)
	testRoom := _testRoom.(int64)

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

	_, err = c.Add(nil, mt.NewGroupTarget(test.G1), testRoom, Live)
	assert.Nil(t, err)

	liveInfo, err := c.FindOrLoadRoom(testRoom)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo)
	assert.Equal(t, testRoom, liveInfo.RoomId)
	assert.Equal(t, "斗鱼官方视频号", liveInfo.RoomName)

	identityInfo, err := c.Get(testRoom)
	assert.Nil(t, err)
	assert.EqualValues(t, liveInfo.GetRoomId(), identityInfo.GetUid())
	assert.EqualValues(t, liveInfo.GetNickname(), identityInfo.GetName())

	liveInfo.ShowStatus = ShowStatus_Living
	liveInfo.VideoLoop = VideoLoopStatus_Off
	liveInfo.liveStatusChanged = true

	testEventChan <- liveInfo

	select {
	case notify := <-testNotifyChan:
		assert.True(t, notify.GetTarget().Equal(mt.NewGroupTarget(test.G1)))
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}

	identityInfo, err = c.Remove(nil, mt.NewGroupTarget(test.G1), testRoom, Live)
	assert.Nil(t, err)
	assert.EqualValues(t, testRoom, identityInfo.GetUid())

	identityInfo, err = c.Remove(nil, mt.NewGroupTarget(test.G1), testRoom, Live)
	assert.NotNil(t, err)
}
