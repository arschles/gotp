package gotp

import (
	"testing"
	"fmt"
	"time"
	"gotp"
)

type TestMessage struct {}

type TestActor struct {
	gotp.GoActor

	Received bool
}

func (t *TestActor) Receive(msg gotp.Message) error {
	fmt.Println("Received message")
	t.Received = true
	return nil
}

func TestActorSpawn(t *testing.T) {
	test := TestActor{Received:false}
	pid := gotp.Spawn(&test)
	pid.Send(TestMessage{})
    time.Sleep(200)
    if &test.Received != true {
    	t.Error("Never received")
    }
}