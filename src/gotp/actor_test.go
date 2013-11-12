package gotp

import (
	"testing"
	"time"
	"fmt"
)

type TestMessage struct {}

type TestActor struct {
	GoActor

	Received int
}

func (t *TestActor) Receive(msg Message) error {
	t.Received++
	return nil
}

func TestActorSpawn(t *testing.T) {
	test := TestActor{Received:0}
	pid := Spawn(&test)
	go func() {
		pid.Send(TestMessage{})
	}()
    time.Sleep(1*time.Second)
    if test.Received != 1 {
    	t.Error("Never received")
    }
}

func BenchmarkActorSingleSender(b *testing.B) {
	test := TestActor{Received:0}
	pid := Spawn(&test)
	for n := 0; n < b.N; n++ {
		pid.Send(TestMessage{})
	}
}

func BenchmarkActorMultiSender(b *testing.B) {
	test := TestActor{Received:0}
	pid := Spawn(&test)
	for n := 0; n < b.N; n++ {
		go func() {
			pid.Send(TestMessage{})
		}()
	}
	for test.Received < b.N {
		fmt.Println(b.N, test.Received)
	}
}