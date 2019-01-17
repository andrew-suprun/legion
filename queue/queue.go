package queue

import (
	"errors"
	"sync"
)

type Queue struct {
	pending []interface{}
	*sync.Cond
	closed bool
}

var Closed = errors.New("Reading from closed queue.")

func New() *Queue {
	return &Queue{Cond: sync.NewCond(&sync.Mutex{})}
}

func (q *Queue) Put(msg interface{}) {
	q.Cond.L.Lock()

	if q.closed {
		q.Cond.L.Unlock()
		return
	}

	q.pending = append(q.pending, msg)
	q.Cond.Signal()
	q.Cond.L.Unlock()
}

func (q *Queue) Get() (msg interface{}) {
	q.Cond.L.Lock()

	for !q.closed && len(q.pending) == 0 {
		q.Cond.Wait()
	}

	if len(q.pending) == 0 && q.closed {
		q.Cond.L.Unlock()
		return Closed
	}

	msg = q.pending[0]
	q.pending = q.pending[1:]
	q.Cond.L.Unlock()
	return msg
}

func (q *Queue) Close() {
	q.Cond.L.Lock()
	q.closed = true
	q.Cond.Broadcast()
	q.Cond.L.Unlock()
}
