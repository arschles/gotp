package gotp

import (
	"testing"
	"time"
	"fmt"
	"runtime"
)

type TestMessage struct {}

type TestActor struct {
	GoActor

	Received int
}

func (t *TestActor) Receive(msg Message) error {
	t.Received++
	if t.Received % 50000 == 0 {
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
		pid.Send(TestMessage{})
	}()
    time.Sleep(1*time.Second)
    if test.Received != 1 {
    	t.Error("Never received")
    }
}

func BenchmarkActorSingleSender(b *testing.B) {
	runtime.GOMAXPROCS(4)
	test := TestActor{Received:0}
	pid := Spawn(&test)
	for n := 0; n < b.N; n++ {
		pid.Send(TestMessage{})
	}
}

func BenchmarkActorMultiSender(b *testing.B) {
	runtime.GOMAXPROCS(4)
	test := TestActor{Received:0}
	pid := Spawn(&test)	
	go func() {
		errChan := pid.Watch()
		err := <- errChan
		fmt.Println("ACTOR ERRORED OUT", err)
	}()
	for n := 0; n < b.N; n++ {
		go func() {
			pid.Send(TestMessage{})
		}()
	    //time.Sleep(1000*time.Nanosecond)
	}
	fmt.Println("Done sending messages")
	for test.Received < b.N {
	}
	pid.Stop()
}