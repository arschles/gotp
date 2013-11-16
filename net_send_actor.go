/*
an actor that can buffer and send messages over the network to a specified hostname and port. example usage:
	//create the actor to buffer 100 messages to localhost:8080, and send all errors to errorPid.
	//this call will attempt to connect to the server and then return the Pid that manages that connection
	sendPid := NetSender(100, "localhost", 8080, errorPid)
	//flush the buffer and send over the wire
	for i := 0; i <= 100; i++ {
		sendPid.Send("hello world")
	}
*/
package gotp

import (
	"encoding/gob"
	"net"
	"time"
)

type NetSendActor struct {
	GoActor
	//the host and port to which this actor will try to send
	host string
	port int
	//the connection. will be nil if there was an error connecting (that error will have been send to the errors Pid)
	conn net.Conn
	//the over-the-wire encoder
	encoder *gob.Encoder
	//this channel will send errors to this pid
	errors Pid
	//the buffer, which will be flushed to the network when it gets full
	buffer []interface{}
}

type NetSendError struct {
	Err error
}

type NetDialError struct {
	Err error
}

type NetSendTimeoutError struct {
	Nanos time.Duration
}

//start the connection to the given host and port. sends all dialing errors to errs.
//once the connection is made, creates and spawns the actor and returns the pid for that actort
func NetSender(bufLen int, h string, p int, errs Pid) *Pid {
	conn, err := net.DialTimeout("tcp", netString(h, p), 500*time.Millisecond)
	if err != nil {
		errs.Send(NetDialError{err})
		return nil
	}
	actor := NetSendActor{host: h, port: p, conn: conn, encoder: gob.NewEncoder(conn), errors: errs, buffer: make([]interface{}, bufLen)}
	pid := Spawn(&actor)
	return &pid
}

func (s *NetSendActor) Receive(msg Message) error {
	if len(s.buffer) < cap(s.buffer) {
		s.buffer = append(s.buffer, msg)
	} else {
		completed := make(chan struct{})
		timeout := make(chan struct{}, 1)
		timeoutNanos := 500 * time.Millisecond
		//the sender goroutine
		go func() {
			err := s.encoder.Encode(s.buffer)
			if err != nil {
				s.errors.Send(NetSendError{err})
				return
			}
			completed <- Unit{}
		}()
		//the timeout goroutine
		go func() {
			time.Sleep(timeoutNanos)
			timeout <- Unit{}
		}()
		select {
		case <-completed:
		case <-timeout:
			s.errors.Send(NetSendTimeoutError{timeoutNanos})
		}

		//regardless of error, delete the buffered elements. the sender handles ack and retry logic
		bufLen := cap(s.buffer)
		s.buffer = make([]interface{}, bufLen)
	}
	return nil
}
