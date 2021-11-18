package utils

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEmitQueue(t *testing.T) {
	c := make(chan *EmitE, 1)
	eq := NewEmitQueue(c, time.Millisecond*100)

	eq.Start()
	defer eq.Stop()
	time.Sleep(time.Millisecond * 500)

	eq.Add(nil)
	eq.Add(NewEmitE(1, test.DouyuLive))
	eq.Add(NewEmitE(2, test.DouyuLive))
	eq.Add(NewEmitE(3, test.DouyuLive))
	for count := 0; count < 3; count++ {
		select {
		case item := <-c:
			assert.EqualValues(t, count+1, item.Id)
			assert.EqualValues(t, test.DouyuLive, item.Type)
		case <-time.After(time.Second * 5):
			assert.Fail(t, "no item received")
		}
	}

	eq.Delete(nil)
	eq.Delete(3)

	for count := 0; count < 3; count++ {
		select {
		case item := <-c:
			assert.NotEqualValues(t, 3, item.Id)
		case <-time.After(time.Second * 5):
			assert.Fail(t, "no item received")
		}
	}

	eq.Update(nil)
	eq.Update(NewEmitE(1, test.BibiliLive))
	eq.Update(NewEmitE(2, test.BibiliLive))

	for count := 0; count < 3; count++ {
		select {
		case item := <-c:
			assert.NotEqualValues(t, 3, item.Id)
			assert.EqualValues(t, test.BibiliLive, item.Type)
		case <-time.After(time.Second * 5):
			assert.Fail(t, "no item received")
		}
	}

	eq.Add(NewEmitE(1, test.BilibiliNews))
	eq.Add(NewEmitE(2, test.BilibiliNews))

	for count := 0; count < 3; count++ {
		select {
		case item := <-c:
			assert.NotEqualValues(t, 3, item.Id)
			assert.EqualValues(t, test.BibiliLive.Add(test.BilibiliNews), item.Type)
		case <-time.After(time.Second * 5):
			assert.Fail(t, "no item received")
		}
	}
}

func TestEmitQueue2(t *testing.T) {
	c := make(chan *EmitE, 100)
	eq := NewEmitQueue(c, time.Millisecond*100)

	eq.Start()
	defer eq.Stop()

	time.Sleep(time.Millisecond * 500)
	eq.Add(NewEmitE(1, test.BibiliLive))
	for count := 0; count < 3; count++ {
		select {
		case item := <-c:
			assert.EqualValues(t, 1, item.Id)
		case <-time.After(time.Second * 5):
			assert.Fail(t, "no item received")
		}
	}

	eq.Delete(1)

	select {
	case <-c:
		assert.Fail(t, "should no item received")
	case <-time.After(time.Second * 1):
	}

LOOP:
	for {
		select {
		case <-c:
		case <-time.After(time.Second * 3):
			break LOOP
		}
	}
}
