package actors

import (
	"fmt"
	"runtime/debug"
	"sync"
)

func NewActor(handler Handler) Actor {
	actor := &actor{
		handler:  handler,
		Cond:     sync.NewCond(&sync.Mutex{}),
		stopChan: make(chan bool, 1),
	}
	go run(actor)
	return actor
}

type Actor interface {
	Send(message interface{})
	Shutdown()
}

var stopMsg = struct{ Stop struct{} }{}

type Handler func(interface{})

type actor struct {
	handler  Handler
	pending  []interface{}
	stopChan chan bool
	*sync.Cond
}

func run(a *actor) {
	for {
		a.Cond.L.Lock()

		if len(a.pending) == 0 {
			a.Cond.Wait()
			a.Cond.L.Unlock()
			continue
		}

		msg := a.pending[0]
		a.pending = a.pending[1:]

		a.Cond.L.Unlock()
		a.handleMessage(msg)

		if msg == stopMsg {
			a.stopChan <- true
			return
		}
	}
}

func (a *actor) handleMessage(msg interface{}) {
	defer func() {
		if r := recover(); r != nil {
			a.Cond.L.Lock()
			fmt.Printf("Actor %v panic: %v\n", a, r)
			fmt.Println(string(debug.Stack()))
			a.Cond.L.Unlock()
		}
	}()
	a.handler(msg)
}

func (a *actor) Send(msg interface{}) {
	a.Cond.L.Lock()
	a.pending = append(a.pending, msg)
	a.Cond.Signal()
	a.Cond.L.Unlock()
}

func (a *actor) Shutdown() {
	a.Send(stopMsg)
	<-a.stopChan
}
