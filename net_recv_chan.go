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
	//the channel to send all errors to, as they happen
	errChan chan error
	//the channel that stops this channel from listening
	stopChan chan struct{}
	//the host and port on which this channel listens
	host string
	port int
}

func (r *RecvChan) asyncSendError(err error) {
	go func() {
		r.errChan <- err
	}()
}

func (r *RecvChan) NetString() string {
	return netString(r.host, r.port)
}

//create a new receive channel. when Start is called on the newly created channel, it will
//listen on TCP on host:port.
//the receiver channel will forward each message it receives on the network to the pid fp
func (r *RecvChan) Init(fp Pid, h string, p int) *RecvChan {
	ec := make(chan error)
	sc := make(chan struct{})
	return &RecvChan{forwardPid: fp, errChan: ec, stopChan: sc, host: h, port: p}
}

//return the channel on which this receiver channel will output its errors
func (r *RecvChan) ErrorChan() chan error {
	return r.errChan
}

//start listening on the network, buffering and forwarding messages
func (r *RecvChan) Start() {
	starterFn := func() {
		listener, err := net.Listen("tcp", r.NetString())
		if err != nil {
			log.Fatalln("error listening", err)
			r.asyncSendError(err)
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
							r.asyncSendError(err)
						} else {
							decoder := gob.NewDecoder(conn)
							decoded := []Unit{}
							err := decoder.Decode(&decoded)
							if err != nil {
								log.Fatalln("error decoding", err)
								r.asyncSendError(err)
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
