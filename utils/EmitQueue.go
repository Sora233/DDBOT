package utils

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	pq "github.com/jupp0r/go-priority-queue"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("emit_queue")

type EmitQueue struct {
	TimeInterval time.Duration

	*sync.Cond
	pq         pq.PriorityQueue
	emitChan   chan<- interface{}
	waitNotify <-chan time.Time
	lastEmit   time.Time
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

		<-q.waitNotify

		e, err := q.pq.Pop()
		if err != nil {
			logger.Errorf("pop from pq failed %v", err)
		} else {
			select {
			case q.emitChan <- e:
			default:
				logger.Errorf("emit chan full")
			}
			q.pq.Insert(e, float64(time.Now().Unix()))
		}
		q.L.Unlock()
	}
}

func (q *EmitQueue) Stop() {

}

func NewEmitQueue(c chan<- interface{}, interval time.Duration) *EmitQueue {
	q := &EmitQueue{
		pq:           pq.New(),
		Cond:         sync.NewCond(new(sync.Mutex)),
		emitChan:     c,
		waitNotify:   time.Tick(interval),
		TimeInterval: interval,
	}
	go q.core()
	return q
}
