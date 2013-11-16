package gotp

import (
	"errors"
	"fmt"
	"time"
)

type Unit struct{}

func makeError(i interface{}) error {
	return errors.New(fmt.Sprintf("%s", i))
}

func forwardPanicToChan(ch chan<- error) {
	if r := recover(); r != nil {
		ch <- makeError(r)
	}
}

type PanicError struct {
	Panic interface{}
}

func forwardPanicToPid(pid Pid) {
	if r := recover(); r != nil {
		pid.Send(PanicError{r})
	}
}

func timeoutChannel(after time.Duration) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		time.Sleep(after)
		ch <- Unit{}
	}()
	return ch
}
