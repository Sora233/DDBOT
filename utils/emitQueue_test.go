package utils

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEmitQueue(t *testing.T) {
	c := make(chan *EmitE, 1)
	eq := NewEmitQueue(c, time.Millisecond*500)

	eq.Start()
	defer eq.Stop()

	eq.Add(NewEmitE(1, test.DouyuLive), time.Unix(10, 0))

	eq.Add(NewEmitE(2, test.DouyuLive), time.Unix(15, 0))

	eq.Add(NewEmitE(3, test.DouyuLive), time.Unix(20, 0))

	for count := 0; count < 3; count++ {
		select {
		case item := <-c:
			assert.EqualValues(t, count+1, item.Id)
		case <-time.After(time.Second * 5):
			assert.Fail(t, "no item received")
		}
	}

}
