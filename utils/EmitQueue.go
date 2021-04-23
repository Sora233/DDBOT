package utils

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/concern"
	pq "github.com/jupp0r/go-priority-queue"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("utils")

type EmitE struct {
	Id   interface{}
	Type concern.Type
}

func NewEmitE(id interface{}, t concern.Type) *EmitE {
	return &EmitE{
		Id:   id,
		Type: t,
	}
}

type EmitQueue struct {
	*sync.Cond

	TimeInterval time.Duration
	pq           pq.PriorityQueue
	emitChan     chan<- *EmitE
	waitTimer    *time.Timer
}

func (q *EmitQueue) Add(e *EmitE, t time.Time) {
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
			if ee, ok := e.(*EmitE); ok {
				select {
				case q.emitChan <- ee:
				default:
					q.L.Unlock()
					q.emitChan <- ee
					q.L.Lock()
				}
			} else {
				logger.Errorf("can not cast type %T", ee)
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

func NewEmitQueue(c chan<- *EmitE, interval time.Duration) *EmitQueue {
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
