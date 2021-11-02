package acfun

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func initConcern(t *testing.T, notifyChan chan<- concern.Notify) *Concern {
	c := NewConcern(notifyChan)
	assert.NotNil(t, c)
	assert.Nil(t, c.Start())
	return c
}

func TestNewConcern(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)

	c := NewConcern(testNotifyChan)
	c.StateManager.UseNotifyGenerator(c.notifyGenerator())
	c.StateManager.UseFreshFunc(func(eventChan chan<- concern.Event) {
		for e := range testEventChan {
			eventChan <- e
		}
	})
	assert.Nil(t, c.StateManager.Start())
	defer c.Stop()
	defer close(testEventChan)

	origUserInfo := &UserInfo{
		Uid:  test.UID1,
		Name: test.NAME1,
	}
	origLiveInfo := &LiveInfo{
		UserInfo: *origUserInfo,
	}

	select {
	case testEventChan <- origLiveInfo:
	default:
		assert.Fail(t, "insert chan failed")
	}

	select {
	case <-testNotifyChan:
		assert.Fail(t, "should no notify received")
	case <-time.After(time.Second):
	}

	_, err := c.StateManager.AddGroupConcern(test.G1, test.UID1, Live)
	assert.Nil(t, err)
	assert.Nil(t, c.StateManager.AddLiveInfo(origLiveInfo))

	select {
	case testEventChan <- origLiveInfo:
	default:
		assert.Fail(t, "insert chan failed")
	}

	select {
	case notify := <-testNotifyChan:
		assert.NotNil(t, notify)
		assert.EqualValues(t, test.UID1, notify.GetUid())
		assert.EqualValues(t, test.G1, notify.GetGroupCode())
	case <-time.After(time.Second):
		assert.Fail(t, "no item received")
	}

	select {
	case testEventChan <- origLiveInfo:
	default:
		assert.Fail(t, "insert chan failed")
	}

	_, err = c.StateManager.AddGroupConcern(test.G2, test.UID1, Live)
	assert.Nil(t, err)
	_, err = c.StateManager.AddGroupConcern(test.G2, test.UID2, Live)
	assert.Nil(t, err)
	for i := 0; i < 2; i++ {
		select {
		case notify := <-testNotifyChan:
			assert.NotNil(t, notify)
			assert.EqualValues(t, test.UID1, notify.GetUid())
			assert.True(t, notify.GetGroupCode() == test.G1 || notify.GetGroupCode() == test.G2)
		case <-time.After(time.Second):
			assert.Fail(t, "no item received")
		}
	}

}
