/**
an abstraction that runs a server on a port and forwards incoming messages to a Pid. example usage:
	//dispatcherPid will receive messages that come over the network, and errorPid will receive all errors that occur.
	//runs the server in the background
	stopChan := StartRecvChan(dispatcherPid, errorPid, "localhost", 8080)
	//do something useful
	stopChan <- Unit{}//blocks until the server has stopped
*/
package gotp

import (
	"encoding/gob"
	"net"
)

type NetListenError struct {
	Err error
}
type NetAcceptError struct {
	Err error
}
type NetDecodeError struct {
	Err error
}

//create a new network server on host:port and start listening on the network in a goroutine.
//all incoming messages will be forwarded to forwardPid and all errors forwarded to errPid.
//send a Unit{} on the returned channel to stop the server. the send will block until the server is stopped
func StartRecvChan(forwardPid Pid, errPid Pid, host string, port int) chan<- struct{} {
	stopChan := make(chan struct{})
	starterFn := func() {
		listener, err := net.Listen("tcp", netString(host, port))
		if err != nil {
			errPid.Send(NetListenError{err})
		} else {
			for {
				defer forwardPanicToPid(errPid)
				select {
				case <-stopChan:
					break
				default:
					{
						conn, err := listener.Accept()
						if err != nil {
							errPid.Send(NetAcceptError{err})
						} else {
							decoder := gob.NewDecoder(conn)
							decoded := []Unit{}
							err := decoder.Decode(&decoded)
							if err != nil {
								errPid.Send(NetDecodeError{err})
							} else {
								forwardPid.Send(decoded)
							}
						}
					}
				}
			}
		}
	}
	go starterFn()
	return stopChan
}
