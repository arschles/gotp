package gotp

import (
	"log"
	"sync"
	"testing"
	"time"
)

const (
	BUFLEN = 100
	HOST   = "localhost"
	PORT   = 8080
)

//the actor that receives messages from a netchan
type NetChanRecv struct {
	GoActor
	Wg sync.WaitGroup
}

func (t *NetChanRecv) Receive(msg Message) error {
	payload := msg.Payload
	log.Println("received", payload)
	switch payload.(type) {
	case int:
		n := payload.(int)
		for i := 0; i < n; i++ {
			t.Wg.Done()
		}
	}
	return nil
}

type ErrorHandler struct {
	GoActor
	NumErrors int
	DevNull   Pid
}

func (e *ErrorHandler) Receive(msg Message) error {
	e.NumErrors++
	e.DevNull.Send(msg)
	return nil
}

func TestSendRecv(t *testing.T) {
	devNullActor := DevNullActor{}
	devNullPid := Spawn(&devNullActor)

	errActor := ErrorHandler{NumErrors: 0, DevNull: devNullPid}
	errPid := Spawn(&errActor)

	//set up the server
	recvActor := NetChanRecv{}
	recvPid := Spawn(&recvActor)
	stopServerChan := StartRecvChan(recvPid, errPid, HOST, PORT)

	//set up the client
	sendPid := NetSender(BUFLEN, HOST, PORT, errPid)

	recvActor.Wg.Add(2)
	sendPid.Send(2)

	wgWaitDone := make(chan Unit)
	wgWaitTimeout := timeoutChannel(500 * time.Millisecond)
	go func() {
		recvActor.Wg.Wait()
		wgWaitDone <- Unit{}
	}()
	select {
	case <-wgWaitDone:
	case <-wgWaitTimeout:
		log.Println("wait group didn't release in time")
		t.Fail()
	}

	serverStopDone := make(chan Unit)
	serverStopTimeout := timeoutChannel(500 * time.Millisecond)
	go func() {
		stopServerChan <- Unit{}
		serverStopDone <- Unit{}
	}()
	select {
	case <-serverStopDone:
	case <-serverStopTimeout:
		log.Println("server didn't stop in time")
		t.Fail()
	}

	//TODO: there's a bad access panic here. fix it
	if errActor.NumErrors > 0 {
		//log.Println("errors was >0")
		t.Fail()
	}
}
