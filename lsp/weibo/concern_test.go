package weibo

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

const testId int64 = 1

func TestConcern(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)

	c := NewConcern(testNotifyChan)

	assert.NotNil(t, c.GetStateManager())

	_, err := c.ParseId("1")
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

	_, err = c.Add(nil, mt.NewGroupTarget(test.G1), testId, News)
	assert.Nil(t, err)

	identityInfo, err := c.Get(testId)
	assert.Nil(t, err)
	assert.EqualValues(t, testId, identityInfo.GetUid())

	newsInfo, err := c.freshNews(testId)
	assert.Nil(t, err)
	assert.NotNil(t, newsInfo)
	newsInfo.Cards = []*Card{
		{CardType: CardType_Normal},
	}

	testEventChan <- newsInfo

	select {
	case notify := <-testNotifyChan:
		assert.True(t, notify.GetTarget().Equal(mt.NewGroupTarget(test.G1)))
		assert.Equal(t, testId, notify.GetUid())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}

	_, err = c.Remove(nil, mt.NewGroupTarget(test.G1), testId, News)
	assert.Nil(t, err)
}
