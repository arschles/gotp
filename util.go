package gotp

import (
	"errors"
	"fmt"
)

type Unit struct{}

func makeError(i interface{}) error {
	return errors.New(fmt.Sprintf("%s", i))
}
