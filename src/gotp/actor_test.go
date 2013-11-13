package gotp

import (
	"testing"
	"fmt"
	"runtime"
	"sync"
	"time"
	"errors"
)

type TestMessage struct {
	id int
}

type TestActor struct {
	GoActor

	Wg sync.WaitGroup
}

func (t *TestActor) Receive(msg Message) error {
	t.Wg.Done()
	if id := msg.Payload.(TestMessage).id; id % 100000 == 0 {
	}
	return nil
}

type DiesActor struct {
	GoActor

	Wg *sync.WaitGroup
}

func (die *DiesActor) Receive(msg Message) error {
	die.Wg.Done()
	return errors.New("I am dead")	
}

type CreatesLinkActor struct {
	GoActor

	Wg sync.WaitGroup
}

func (link *CreatesLinkActor) Receive(msg Message) error {

	toStart := DiesActor{Wg: &link.Wg}
	pid := link.StartLink(&toStart)
	pid.Send(Unit{})
	return nil
}

type CreatesChildActor struct {
	GoActor
	child Pid
	Wg sync.WaitGroup
}

func (ch *CreatesChildActor) Receive(msg Message) error {
	switch msg.Payload.(type) {
	case Stop:
		ch.Wg.Done()
	default:
		toStart := DiesActor{Wg: &ch.Wg}
		ch.child = ch.StartChild(&toStart)
		ch.child.Send(Unit{})
	}
	return nil
}

func TestActorSpawn(t *testing.T) {
	runtime.GOMAXPROCS(4)
	test := TestActor{}
	test.Wg.Add(1)
	pid := Spawn(&test)
	go func() {
		pid.Send(TestMessage{1})
	}()
    test.Wg.Wait()
    pid.Stop()
}

func TestStartLink(t *testing.T) {
	runtime.GOMAXPROCS(4)
	test := CreatesLinkActor{}
	test.Wg.Add(1)
	pid := Spawn(&test)
	go func() {
		pid.Send(Unit{})
	}()
	test.Wg.Wait()
	time.Sleep(1*time.Second)
	if test.Running() {
		t.Error("Actor still running")
	}
	pid.Stop()
}

func TestStartChild(t *testing.T) {
	runtime.GOMAXPROCS(4)
	test := CreatesChildActor{}
	test.Wg.Add(2)
	pid := Spawn(&test)
	go func() {
		pid.Send(Unit{})
	}()
	test.Wg.Wait()
	pid.Stop()
}

func BenchmarkActorSingleSender(b *testing.B) {
	runtime.GOMAXPROCS(4)
	test := TestActor{}
	test.Wg.Add(b.N)
	pid := Spawn(&test)
	go func() {
		for n := 0; n < b.N; n++ {
			pid.Send(TestMessage{n})
		}
	}()
	test.Wg.Wait()
	pid.Stop()
}

func BenchmarkActorTenSenders(b *testing.B) {
	runtime.GOMAXPROCS(4)
	test := TestActor{}
	test.Wg.Add(b.N*10)
	pid := Spawn(&test)
	for i := 0; i < 10; i++ {
		go func() {
			for n := 0; n < b.N; n++ {
				pid.Send(TestMessage{n})
			}
		}()
	}
	test.Wg.Wait()
	pid.Stop()
}

func BenchmarkActorMultiSender(b *testing.B) {
	runtime.GOMAXPROCS(4)
	test := TestActor{}
	test.Wg.Add(b.N)
	pid := Spawn(&test)
	go func() {
		errChan := pid.Watch()
		err := <- errChan
		fmt.Println("ACTOR ERRORED OUT", err)
	}()
	for n := 0; n < b.N; n++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Failed to send", r)
				}
			}()
			pid.Send(TestMessage{n})
		}()
	}
	test.Wg.Wait()
	pid.Stop()
}

func BenchmarkBaseline(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	runtime.GOMAXPROCS(4)
	channel := make(chan int)
	go func() {
		for n := 0; n < b.N; n++ {
			channel <- n
		}
	}()
	go func() {
		for n := 0; n < b.N; n++ {
			<- channel
			wg.Done()
		}
	}()
	wg.Wait()
}