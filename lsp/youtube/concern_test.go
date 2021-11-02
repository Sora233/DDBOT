package youtube

import (
	"context"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConcern(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)

	c := NewConcern(testNotifyChan)

	assert.NotNil(t, c.GetStateManager())

	c.StateManager.UseNotifyGenerator(c.notifyGenerator())
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

	_, err := c.ParseId(test.NAME1)
	assert.Nil(t, err)

	err = c.StateManager.AddInfo(&Info{
		UserInfo: UserInfo{
			ChannelId:   test.NAME1,
			ChannelName: test.NAME1,
		},
	})
	assert.Nil(t, err)

	_, err = c.StateManager.AddGroupConcern(test.G1, test.NAME1, Live)
	assert.Nil(t, err)

	identityInfo, err := c.Get(test.NAME1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.NAME1, identityInfo.GetUid())

	assert.NotNil(t, c.GetGroupConcernConfig(test.G1, test.NAME1))

	identityInfos, _, err := c.List(test.G1, Live)
	assert.Nil(t, err)
	assert.Len(t, identityInfos, 1)

	testEventChan <- &VideoInfo{
		UserInfo: UserInfo{
			ChannelId:   test.NAME1,
			ChannelName: test.NAME1,
		},
		VideoType:         VideoType_Live,
		VideoStatus:       VideoStatus_Living,
		LiveStatusChanged: true,
	}

	testEventChan <- &VideoInfo{
		UserInfo: UserInfo{
			ChannelId:   test.NAME1,
			ChannelName: test.NAME1,
		},
		VideoType:         VideoType_Live,
		VideoStatus:       VideoStatus_Living,
		LiveStatusChanged: true,
	}

	select {
	case notify := <-testNotifyChan:
		assert.Equal(t, test.G1, notify.GetGroupCode())
		assert.Equal(t, test.NAME1, notify.GetUid())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}

	select {
	case <-testNotifyChan:
		assert.Fail(t, "should no notify received")
	case <-time.After(time.Second):

	}

	_, err = c.Remove(nil, test.G1, test.NAME1, Live)
	assert.Nil(t, err)
}
