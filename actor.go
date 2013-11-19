package gotp

import (
	"sync"
)

//public types

//the actor itself simply must define these functions
//DON'T IMPLEMENT THIS DIRECTLY, Implement GoActor
type Actor interface {
	Receive(msg Message) error

	Init()
	Self() Pid
	selfInit(self Pid)

	StartChild(actor Actor) Pid
	StartLink(actor Actor) Pid
	stop()
}

type GoActorWithInit struct {
	self  Pid
	alive bool
}

func (ac *GoActorWithInit) selfInit(pid Pid) {
	ac.self = pid
	ac.alive = true
}

func (ac *GoActorWithInit) Self() Pid {
	return ac.self
}

func (ac *GoActorWithInit) StartChild(actor Actor) Pid {
	childPid := Spawn(actor)
	watch := childPid.Watch()
	go func() {
		err := <-watch
		ac.self.recv <- Message{Payload: Stop{err: err, Sender: childPid}}
	}()
	return childPid
}

func (ac *GoActorWithInit) StartLink(actor Actor) Pid {
	linkPid := Spawn(actor)
	watch := linkPid.Watch()
	go func() {
		err := <-watch
		ac.self.stop <- err
	}()
	return linkPid
}

func (ac *GoActorWithInit) stop() {
	ac.alive = false
}

type GoActor struct {
	GoActorWithInit
}

func (ac *GoActor) Init() {}

type Stop struct {
	err    error
	Sender Pid
}

//send a message asynchronously to the pid
func (p *Pid) Send(msg interface{}) Unit {
	m := Message{msg}
	p.recv <- m
	return Unit{}
}

//begin the shutdown process of pid, and send on the returned channel when the shutdown finished
func (p *Pid) Stop() chan Unit {
	stopped := make(chan Unit)
	go func() {
		p.stop <- nil
		stopped <- Unit{}
	}()
	return stopped
}

//watch a pid for errors, and send on the returned channel if an error occured
func (p *Pid) Watch() chan error {
	errChan := make(chan error)
	go func() {
		err := <-p.errored
		errChan <- err
	}()
	return errChan
}

//create a new actor and return the pid, so you can send it messages
func Spawn(actor Actor) Pid {
	p := Pid{recv: make(chan Message), stop: make(chan error), errored: make(chan error)}
	//start the receive loop
	go func() {
		actor.selfInit(p)
		actor.Init()
		recvLoop(p.recv, p, actor)
	}()
	return p
}

//create a new buffered actor and return the pid; a buffered actor has a limited mailbox,
//subsequent messages to the buffered actor when mailbox is full will block
func SpawnBuffered(actor Actor, buffer int) Pid {
	p := Pid{recv: make(chan Message, buffer), stop: make(chan error), errored: make(chan error)}
	go func() {
		actor.selfInit(p)
		actor.Init()
		simpleRecvLoop(p.recv, p, actor)
	}()
	return p
}

func recvLoop(recv chan Message, p Pid, actor Actor) {
	//create the first nextwait channel
	nextWait := make(chan bool)
	firstWait := nextWait
	go func() {
		firstWait <- true
	}()
	//handle panics in this loop
	defer forwardPanicToChan(p.errored)

	goingLock := sync.Mutex{}
	//loop forever
	going := true

	stillGoing := func() bool {
		goingLock.Lock()
		defer goingLock.Unlock()
		return going
	}

	stopGoing := func() {
		goingLock.Lock()
		going = false
		actor.stop()
		goingLock.Unlock()
	}

	go func() {
		<-p.stop
		stopGoing()
	}()

	for stillGoing() {
		received := <-p.recv
		if going == false {
			return
		}
		currWait := nextWait
		nextWait = make(chan bool)
		opsNextWait := nextWait
		runFn := func() {
			defer func() {
				if r := recover(); r != nil {
					p.errored <- makeError(r)
					stopGoing()
				}
			}()
			<-currWait
			err := actor.Receive(received)
			if err != nil {
				p.errored <- err
				stopGoing()
			}
			opsNextWait <- true
		}
		go runFn()
	}
}

func simpleRecvLoop(recv chan Message, p Pid, actor Actor) {
	//handle panics in this loop
	defer forwardPanicToChan(p.errored)

	for {
		received := <-p.recv
		err := actor.Receive(received)
		if err != nil {
			p.errored <- err
			return
		}
	}

}
