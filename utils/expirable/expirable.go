package expirable

import (
	"go.uber.org/atomic"
	"sync"
	"time"
)

type Expirable struct {
	d        time.Duration
	deadline atomic.Pointer[time.Time]
	mu       sync.Mutex
	val      atomic.Value
	f        func() interface{}
}

func (e *Expirable) Do() interface{} {
	now := time.Now()
	if e.deadline.Load().Before(now) {
		e.mu.Lock()
		if e.deadline.Load().Before(now) {
			e.val.Store(e.f())
			newDdl := now.Add(e.d)
			e.deadline.Store(&newDdl)
		}
		e.mu.Unlock()
	}
	return e.val.Load()
}

func NewExpirable(duration time.Duration, action func() interface{}) *Expirable {
	e := &Expirable{
		f: action,
		d: duration,
	}
	e.deadline.Store(&time.Time{})
	return e
}
