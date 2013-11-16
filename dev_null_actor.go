package gotp

import "log"

type DevNullActor struct {
	GoActor
}

func (d *DevNullActor) Receive(msg Message) error {
	log.Println("dev-null-actor", "received", msg)
	return nil
}
