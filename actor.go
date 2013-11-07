package goactor

import (
	"container/list"
	"errors"
	"sync"
)

//public types

//the actor itself simply must define these functions
type Actor interface {
	//Passing in the pid here allows us to call self.StartChild, self.StartLink, etc
	Recieve(msg Message, self Pid) Unit
}

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

//start a child of the given Pid
func (p Pid) StartChild(actor Actor) Pid {
	//for now just spawn, in the future wire up watches
	child := Spawn(actor)
	return child
}

//create a new actor and return the pid, so you can send it messages
func Spawn(actor Actor) Pid {
	p := Pid{queue: list.New(), queue_lock: new(sync.RWMutex), ready: make(chan Unit)}
	ready := make(chan Unit)
	//start the receive loop
	go recvLoop(ready, p, actor)
	return p
}

//run a receive loop
func recvLoop(ready chan Unit, p Pid, actor Actor) {
	select {
		case <- p.ready : {
			p.queue_lock.RLock()
			elt := p.queue.Front()
			if elt != nil {
				p.queue.Remove(elt)
				actor.Receive(elt.Value.(Message), p)
			}
			p.queue_lock.RUnlock()
			recvLoop(ready, p, actor)
		}
		case <- p.stop : {
			return
		}
	}
}
