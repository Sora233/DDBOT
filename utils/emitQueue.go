package utils

import (
	"container/list"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"go.uber.org/atomic"
	"sync"
	"time"
)

type EmitE struct {
	Id   interface{}
	Type concern_type.Type
}

func NewEmitE(id interface{}, t concern_type.Type) *EmitE {
	return &EmitE{
		Id:   id,
		Type: t,
	}
}

type EmitQueue struct {
	TimeInterval time.Duration

	stopped   atomic.Bool
	stop      chan interface{}
	eqlist    *list.List
	eqlistCur *list.Element
	emitChan  chan<- *EmitE
	waitTimer *time.Timer
	cond      *sync.Cond
	wg        sync.WaitGroup
}

func (q *EmitQueue) Add(e *EmitE) {
	if e == nil {
		return
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.eqlist.Len() == 0 {
		q.eqlist.PushBack(e)
	} else {
		var found = false
		for head := q.eqlist.Front(); head != nil; head = head.Next() {
			if headE := head.Value.(*EmitE); headE.Id == e.Id {
				headE.Type = headE.Type.Add(e.Type)
				found = true
				break
			}
		}
		if !found {
			q.eqlist.PushBack(e)
		}
	}
	if q.eqlist.Len() == 1 {
		q.eqlistCur = q.eqlist.Front()
		q.cond.Signal()
	}
}

func (q *EmitQueue) Update(e *EmitE) {
	if e == nil {
		return
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for head := q.eqlist.Front(); head != nil; head = head.Next() {
		headE := head.Value.(*EmitE)
		if headE.Id != e.Id {
			continue
		}
		headE.Type = e.Type
	}
}

func (q *EmitQueue) Delete(id interface{}) {
	if id == nil {
		return
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for head := q.eqlist.Front(); head != nil; head = head.Next() {
		headE := head.Value.(*EmitE)
		if headE.Id != id {
			continue
		}
		if q.eqlistCur == head {
			q.eqlistCur = q.eqlistCur.Next()
		}
		q.eqlist.Remove(head)
	}
	if q.eqlist.Len() == 0 {
		q.eqlistCur = nil
	}
}

func (q *EmitQueue) core() {
	q.wg.Add(1)
	defer q.wg.Done()
	for {
		q.cond.L.Lock()
		for q.eqlist.Len() == 0 && !q.stopped.Load() {
			q.cond.Wait()
		}
		q.cond.L.Unlock()

		select {
		case <-q.waitTimer.C:
		case <-q.stop:
			return
		}
		q.cond.L.Lock()
		if q.eqlist.Len() > 0 {
			if q.eqlistCur == nil {
				q.eqlistCur = q.eqlist.Front()
			}
			headE := q.eqlistCur.Value.(*EmitE)
			q.eqlistCur = q.eqlistCur.Next()
			q.cond.L.Unlock()

			select {
			case q.emitChan <- headE:
			case <-q.stop:
				return
			}
		} else {
			q.cond.L.Unlock()
		}
		q.waitTimer.Reset(q.TimeInterval)
	}
}

func (q *EmitQueue) Start() {
	go q.core()
}

func (q *EmitQueue) Stop() {
	q.stopped.Store(true)
	if q.stop != nil {
		close(q.stop)
	}
	q.cond.Broadcast()
	q.wg.Wait()
}

func NewEmitQueue(c chan<- *EmitE, interval time.Duration) *EmitQueue {
	q := &EmitQueue{
		stop:         make(chan interface{}),
		cond:         sync.NewCond(new(sync.Mutex)),
		eqlist:       list.New(),
		emitChan:     c,
		waitTimer:    time.NewTimer(interval),
		TimeInterval: interval,
	}
	return q
}
