package gotp

import "fmt"

type WireMessage struct {
	destination Pid
	message     Message
}

func netString(host string, port int) string {
	return fmt.Sprintf("%v:%v", host, port)
}
