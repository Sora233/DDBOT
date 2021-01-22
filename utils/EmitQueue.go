package utils

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	pq "github.com/jupp0r/go-priority-queue"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("utils")

type EmitQueue struct {
	*sync.Cond

	TimeInterval time.Duration
	pq           pq.PriorityQueue
	emitChan     chan<- interface{}
	waitTimer    *time.Timer
}

func (q *EmitQueue) Add(e interface{}, t time.Time) {
	q.L.Lock()
	defer q.L.Unlock()
	q.pq.Insert(e, float64(t.Unix()))
	if q.pq.Len() == 1 {
		q.Signal()
	}
}

func (q *EmitQueue) core() {
	for {
		q.L.Lock()
		for q.pq.Len() == 0 {
			q.Wait()
		}
		q.L.Unlock()

		<-q.waitTimer.C

		q.L.Lock()
		e, err := q.pq.Pop()
		if err != nil {
			logger.Errorf("pop from pq failed %v", err)
			q.L.Unlock()
		} else {
			select {
			case q.emitChan <- e:
			default:
				q.L.Unlock()
				q.emitChan <- e
				q.L.Lock()
			}

			q.pq.Insert(e, float64(time.Now().Unix()))
			q.L.Unlock()
		}
		q.waitTimer.Reset(q.TimeInterval)
	}
}

func (q *EmitQueue) Stop() {
	// TODO
}

func NewEmitQueue(c chan<- interface{}, interval time.Duration) *EmitQueue {
	q := &EmitQueue{
		pq:           pq.New(),
		Cond:         sync.NewCond(new(sync.Mutex)),
		emitChan:     c,
		waitTimer:    time.NewTimer(interval),
		TimeInterval: interval,
	}
	go q.core()
	return q
}
