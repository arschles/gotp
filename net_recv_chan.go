/**
an abstraction that runs a server on a port and forwards incoming messages to a Pid
*/
package gotp

import (
	"encoding/gob"
	"log"
	"net"
)

//a type that represents a network stack that waits for messages and forwards them to a pid. all errors get forwarded to errChan
type RecvChan struct {
	//the Pid to foward messages to. this Pid could be a dispatcher to route messages to the right place
	forwardPid Pid
	//the Pid to forward errors to
	errPid Pid
	//the channel that stops this channel from listening
	stopChan chan struct{}
	//the host and port on which this channel listens
	host string
	port int
}

type NetListenError struct {
	Err error
}
type NetAcceptError struct {
	Err error
}
type NetDecodeError struct {
	Err error
}

func (r *RecvChan) NetString() string {
	return netString(r.host, r.port)
}

//create a new receive channel. when Start is called on the newly created channel, it will
//listen on TCP on host:port.
//the receiver channel will forward each message it receives on the network to the pid fp,
//and it will forward each error it encounters to ep
func (r *RecvChan) Init(fp Pid, ep Pid, h string, p int) *RecvChan {
	sc := make(chan struct{})
	return &RecvChan{forwardPid: fp, errPid: ep, stopChan: sc, host: h, port: p}
}

//start listening on the network, buffering and forwarding messages
func (r *RecvChan) Start() {
	starterFn := func() {
		listener, err := net.Listen("tcp", r.NetString())
		if err != nil {
			log.Fatalln("error listening", err)
			r.errPid.Send(NetListenError{err})
		} else {
			for {
				select {
				case <-r.stopChan:
					break
				default:
					{
						conn, err := listener.Accept()
						if err != nil {
							log.Fatalln("error accepting", err)
							r.errPid.Send(NetAcceptError{err})
						} else {
							decoder := gob.NewDecoder(conn)
							decoded := []Unit{}
							err := decoder.Decode(&decoded)
							if err != nil {
								log.Fatalln("error decoding", err)
								r.errPid.Send(NetDecodeError{err})
							} else {
								r.forwardPid.Send(decoded)
							}
						}
					}
				}
			}
		}
	}
	go starterFn()
}

//stop listening on the network. blocks until the stop is complete
func (r *RecvChan) Stop() {
	r.stopChan <- Unit{}
}
