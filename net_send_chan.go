package gotp

import (
	"encoding/gob"
	"log"
	"net"
)

type SendChan struct {
	Host string
	Port int
}

//start the send channel and return 3 channels to communicate the state of the send channel:
//- if you receive on dialError, do not receive or send on any other channel because this send channel is broken. try calling Start again
//- otherwise, you can send on sendChannel. this channel is buffered to bufSize. be aware that sends on this channel may block if a network send happens when the next buffer is flushed
//- if you receive on sendError, a buffer flush failed. regardless of whether the buffered elements made it over the wire, the buffer was flushed
func (s *SendChan) Start(bufLen int) (dialError <-chan error, sendChannel chan<- interface{}, sendError <-chan error) {
	dialErrChan := make(chan error)
	sendChan := make(chan interface{}, bufLen)
	sendErrChan := make(chan error)

	dialAndListen := func() {
		conn, err := net.Dial("tcp", netString(s.Host, s.Port))
		if err != nil {
			log.Fatalln("error dialing", err)
			dialErrChan <- err
		} else {
			encoder := gob.NewEncoder(conn)
			buffer := make([]interface{}, bufLen)
			for {
				buffered := <-sendChan
				if len(buffer) < bufLen {
					//append to the buffer if it's not full
					buffer = append(buffer, buffered)
				} else {
					//flush the buffer if it's full
					err := encoder.Encode(buffer)
					if err != nil {
						log.Fatalln("error encoding and sending", err)
						sendErrChan <- err
					}
					//even if there was an error, delete the buffered elements, so the sender can retry
					buffer = make([]interface{}, bufLen)
				}
			}
		}
	}
	go dialAndListen()
	return dialErrChan, sendChan, sendErrChan
}
