package gotp

import (
	"testing"
	"time"
	"fmt"
	"runtime"
)

type TestMessage struct {
	id int
}

type TestActor struct {
	GoActor

	Received int
}

func (t *TestActor) Receive(msg Message) error {
	t.Received++
	if t.Received % 100000 == 0 {
		fmt.Println(t.Received, runtime.NumGoroutine())
	}
	return nil
}

func TestActorSpawn(t *testing.T) {
	runtime.GOMAXPROCS(4)
	test := TestActor{Received:0}
	pid := Spawn(&test)
	go func() {
		errChan := pid.Watch()
		err := <- errChan
		fmt.Println("ACTOR ERRORED OUT", err)
	}()
	go func() {
		pid.Send(TestMessage{1})
	}()
    time.Sleep(1*time.Second)
    if test.Received != 1 {
    	t.Error("Never received")
    }
}

func BenchmarkActorSingleSender(b *testing.B) {
	runtime.GOMAXPROCS(2)
	test := TestActor{Received:0}
	pid := Spawn(&test)
	go func() {
		for n := 0; n < b.N; n++ {
			pid.Send(TestMessage{n})
		}
	}()
	fmt.Println("Done sending messages")
	for test.Received < b.N {
	}
	pid.Stop()
}

func BenchmarkActorMultiSender(b *testing.B) {
	runtime.GOMAXPROCS(32)
	test := TestActor{Received:0}
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
	    //time.Sleep(1000*time.Nanosecond)
	}
	fmt.Println("Done sending messages")
	for test.Received < b.N {
	}
	pid.Stop()
}