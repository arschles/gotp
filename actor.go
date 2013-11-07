package goactor

import (
	"container/list"
	"errors"
	"sync"
)

//public types
type Message struct {
	Sender  Pid
	Payload interface{}
}

type Unit struct{}

type Pid struct {
	//the queue of messages for this Pid, and the read/write lock to protect it
	queue      *list.List
	queue_lock *sync.RWMutex
	//the channel to signal that the queue is ready for reading
	ready      chan Unit
	//the channel to signal that the actor backing this pid should shut down
	stop       chan Unit
}

type Receiver func(msg Message) Unit

//send a message asynchronously to the pid
func (p Pid) Send(msg interface{}) Unit {
	m := Message{Sender: p, Payload: msg}
	p.queue_lock.Lock()
	p.queue.PushBack(m)
	go func() {
		p.ready <- Unit{}
	}()
	p.queue_lock.Unlock()
	return Unit{}
}

//begin the shutdown process of pid. the returned channel will finish when the shutdown finishes
func (p Pid) Stop() chan Unit {
	stopped := make(chan Unit)
	go func() {
		p.stop <- Unit{}
		stopped <- Unit{}
	}
	return stopped
}

//create a new actor and return the pid, so you can send it messages
func Spawn(fn Receiver) Pid {
	p := Pid{queue: list.New(), queue_lock: new(sync.RWMutex), ready: make(chan Unit)}
	ready := make(chan Unit)
	//start the receive loop
	go recvLoop(ready, p, fn)
	return p
}

//run a receive loop
func recvLoop(ready chan Unit, p Pid, fn Receiver) {
	select {
		case <- p.ready : {
			p.queue_lock.RLock()
			elt := p.queue.Front()
			if elt != nil {
				p.queue.Remove(elt)
				fn(elt.Value.(Message))
			}
			p.queue_lock.RUnlock()
			recvLoop(ready, p, fn)
		}
		case <- p.stop : {
			return
		}
	}
}
