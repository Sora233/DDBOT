package acfun

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

var (
	g1 = mt.NewGroupTarget(test.G1)
	g2 = mt.NewGroupTarget(test.G2)
)

func TestNewConcern(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)

	c := NewConcern(testNotifyChan)
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
	assert.NotNil(t, c.GetStateManager())
	assert.Nil(t, c.StateManager.Start())
	defer c.Stop()
	defer close(testEventChan)

	_id, err := c.ParseId("123")
	assert.Nil(t, err)
	assert.EqualValues(t, 123, _id)

	origUserInfo := &UserInfo{
		Uid:  test.UID1,
		Name: test.NAME1,
	}
	origLiveInfo := &LiveInfo{
		UserInfo:          *origUserInfo,
		IsLiving:          true,
		liveStatusChanged: true,
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

	_, err = c.StateManager.AddTargetConcern(g1, test.UID1, Live)
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
		assert.True(t, notify.GetTarget().Equal(g1))
	case <-time.After(time.Second):
		assert.Fail(t, "no item received")
	}

	_, err = c.StateManager.AddTargetConcern(g2, test.UID1, Live)
	assert.Nil(t, err)
	_, err = c.StateManager.AddTargetConcern(g2, test.UID2, Live)
	assert.Nil(t, err)
	err = c.StateManager.AddUserInfo(&UserInfo{
		Uid:  test.UID2,
		Name: test.NAME2,
	})
	assert.Nil(t, err)

	select {
	case testEventChan <- origLiveInfo:
	default:
		assert.Fail(t, "insert chan failed")
	}

	for i := 0; i < 2; i++ {
		select {
		case notify := <-testNotifyChan:
			assert.NotNil(t, notify)
			assert.EqualValues(t, test.UID1, notify.GetUid())
			assert.True(t, notify.GetTarget().Equal(g1) || notify.GetTarget().Equal(g2))
		case <-time.After(time.Second):
			assert.Fail(t, "no item received")
		}
	}

	go c.TargetWatchNotify(g2, test.UID1)
	select {
	case notify := <-testNotifyChan:
		assert.NotNil(t, notify)
		assert.EqualValues(t, test.UID1, notify.GetUid())
		assert.True(t, notify.GetTarget().Equal(g2))
		assert.NotNil(t, notify.Logger())
		assert.NotNil(t, notify.ToMessage())
	case <-time.After(time.Second):
		assert.Fail(t, "no item received")
	}

}

func TestNewConcern2(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testNotifyChan := make(chan concern.Notify)
	c := NewConcern(testNotifyChan)
	assert.Nil(t, c.Start())
	defer c.Stop()

	timeup := time.After(time.Second * 5)

	const testId int64 = 1

	info, err := c.Add(nil, g1, testId, Live)
	assert.Nil(t, err)
	assert.EqualValues(t, "admin", info.GetName())
	assert.EqualValues(t, testId, info.GetUid())

	info, err = c.Remove(nil, g1, testId, Live)
	assert.Nil(t, err)
	assert.EqualValues(t, "admin", info.GetName())
	assert.EqualValues(t, testId, info.GetUid())

	<-timeup

}
