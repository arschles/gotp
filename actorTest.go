package _gotp

import (
	"testing"
	"fmt"
	"time"
)

type TestMessage struct {}

type TestActor struct {
	GoActor

	Received := false

	func Receive(msg Message) error {
		Received = true
	}

}

func TestActorSpawn(t *testing.T) {
	test := TestActor{}
	pid := Spawn(test)
	pid.Send(TestMessage)
    time.Sleep(200)
    if &test.Received != true {
    	t.Error("Never received")
    }
}