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
	TimeInterval time.Duration
	pq           pq.PriorityQueue
	emitChan     chan<- *EmitE
	waitTimer    *time.Timer
	cond         *sync.Cond
}

func (q *EmitQueue) Add(e *EmitE, t time.Time) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.pq.Insert(e, float64(t.Unix()))
	if q.pq.Len() == 1 {
		q.cond.Signal()
	}
}

func (q *EmitQueue) core() {
	for {
		q.cond.L.Lock()
		for q.pq.Len() == 0 {
			q.cond.Wait()
		}
		q.cond.L.Unlock()

		<-q.waitTimer.C

		q.cond.L.Lock()
		e, err := q.pq.Pop()
		if err != nil {
			logger.Errorf("pop from pq failed %v", err)
			q.cond.L.Unlock()
		} else {
			if ee, ok := e.(*EmitE); ok {
				select {
				case q.emitChan <- ee:
				default:
					q.cond.L.Unlock()
					q.emitChan <- ee
					q.cond.L.Lock()
				}
			} else {
				logger.Errorf("can not cast type %T", ee)
			}

			q.pq.Insert(e, float64(time.Now().Unix()))
			q.cond.L.Unlock()
		}
		q.waitTimer.Reset(q.TimeInterval)
	}
}

func (q *EmitQueue) Start() {
	go q.core()
}

func (q *EmitQueue) Stop() {
	// TODO
}

func NewEmitQueue(c chan<- *EmitE, interval time.Duration) *EmitQueue {
	q := &EmitQueue{
		pq:           pq.New(),
		cond:         sync.NewCond(new(sync.Mutex)),
		emitChan:     c,
		waitTimer:    time.NewTimer(interval),
		TimeInterval: interval,
	}
	return q
}
