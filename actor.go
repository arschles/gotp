package goactor

import (
	"container/list"
	"errors"
	"fmt"
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
	ready chan Unit
	//the channel to signal that the actor backing this pid should shut down
	stop chan Unit
	//the channel to signal a watcher that the actor backing this pid errored
	errored chan error
}

type Receiver func(msg Message) error

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

//begin the shutdown process of pid, and send on the returned channel when the shutdown finished
func (p Pid) Stop() chan Unit {
	stopped := make(chan Unit)
	go func() {
		p.stop <- Unit{}
		stopped <- Unit{}
	}()
	return stopped
}

//watch a pid for errors, and send on the returned channel if an error occured
func (p Pid) Watch() chan error {
	errChan := make(chan error)
	go func() {
		err := <-p.errored
		errChan <- err
	}()
	return errChan
}

//create a new actor and return the pid, so you can send it messages
func Spawn(fn Receiver) Pid {
	p := Pid{queue: list.New(), queue_lock: new(sync.RWMutex), ready: make(chan Unit), stop: make(chan Unit), errored: make(chan error)}
	ready := make(chan Unit)
	//start the receive loop
	go recvLoop(ready, p, fn)
	return p
}

func makeError(i interface{}) error {
	return errors.New(fmt.Sprintf("%s", i))
}

//run a receive loop
func recvLoop(ready chan Unit, p Pid, fn Receiver) {
	select {
	case <-p.ready:
		{
			p.queue_lock.RLock()
			elt := p.queue.Front()
			if elt != nil {
				p.queue.Remove(elt)
				defer func() {
					if r := recover(); r != nil {
						go func() {
							p.errored <- makeError(r)
						}()
					}
				}()
				err := fn(elt.Value.(Message))
				if err != nil {
					p.errored <- err
				}
			}
			p.queue_lock.RUnlock()
			recvLoop(ready, p, fn)
		}
	case <-p.stop:
		{
			return
		}
	}
}
