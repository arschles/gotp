package gotp

import (
	"errors"
	"fmt"
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
