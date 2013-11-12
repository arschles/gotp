package gotp

import (
	"testing"
	"fmt"
	"time"
)

type TestMessage struct {}

type TestActor struct {
	GoActor

	Received bool
}

func (t *TestActor) Receive(msg Message) error {
	fmt.Println("Received message")
	t.Received = true
	return nil
}

func TestActorSpawn(t *testing.T) {
	test := TestActor{Received:false}
	pid := Spawn(&test)
	fmt.Println("Sending message")
	pid.Send(TestMessage{})
    time.Sleep(200)
    if test.Received != true {
    	t.Error("Never received")
    }
}