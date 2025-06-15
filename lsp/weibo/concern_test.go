package weibo

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
	"github.com/Sora233/DDBOT/v2/lsp/concern"
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

	_, err = c.Add(nil, test.G1, testId, News)
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
		assert.Equal(t, test.G1, notify.GetGroupCode())
		assert.Equal(t, testId, notify.GetUid())
	case <-time.After(time.Second):
		assert.Fail(t, "no notify received")
	}

	_, err = c.Remove(nil, test.G1, testId, News)
	assert.Nil(t, err)
}
