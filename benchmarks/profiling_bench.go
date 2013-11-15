package main

import (
	"sync"
	"log"
	"net/http"
	"gotp"
	"fmt"
)

import _ "net/http/pprof"

type TestMessage struct {
	id int
}

//A good example actor
type TestActor struct {
	gotp.GoActor

	Wg sync.WaitGroup
}

func (t *TestActor) Receive(msg gotp.Message) error {
	t.Wg.Done()
	return nil
}

//for profiling
func main() {
	runs := 1000000000
	fmt.Println("Starting", runs)
	go func() {
		log.Println(http.ListenAndServe("localhost:6161", nil))
	}()
	test := TestActor{}
	test.Wg.Add(runs)
	pid := gotp.Spawn(&test)
	go func() {
		for n := 0; n < runs; n++ {
			pid.Send(TestMessage{n})
		}
	}()
	test.Wg.Wait()
	pid.Stop()
}