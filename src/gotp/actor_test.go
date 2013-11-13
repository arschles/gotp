package gotp

import (
	"testing"
	"fmt"
	"runtime"
	"sync"
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

func TestActorSpawn(t *testing.T) {
	runtime.GOMAXPROCS(4)
	test := TestActor{}
	test.Wg.Add(1)
	pid := Spawn(&test)
	go func() {
		errChan := pid.Watch()
		err := <- errChan
		fmt.Println("ACTOR ERRORED OUT", err)
	}()
	go func() {
		pid.Send(TestMessage{1})
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

func BenchmarkGoChannels(b *testing.B) {
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