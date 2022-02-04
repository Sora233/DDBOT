package expirable

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"testing"
	"time"
)

func TestExpirable(t *testing.T) {
	var counter = atomic.NewInt64(0)

	e := NewExpirable(time.Second, func() interface{} {
		return counter.Inc()
	})

	timer := time.NewTimer(time.Millisecond * 500)

F:
	for {
		select {
		case <-timer.C:
			break F
		default:
		}
		c := e.Do()
		assert.EqualValues(t, 1, c)
		time.Sleep(time.Millisecond * 20)
	}

	time.Sleep(time.Second)

	c := e.Do()
	assert.EqualValues(t, 2, c)
}
