package gotp

import (
	"sync"
	"testing"
)

type TestGenServer struct {
	Wg *sync.WaitGroup
}

func (t *TestGenServer) Init() {}

func (t *TestGenServer) HandleCast(msg Message) {
	t.Wg.Done()
}

func (t *TestGenServer) HandleCall(msg Message) interface{} {
	defer func() {
		t.Wg.Done()
	}()
	return msg
}

func TestCast(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	pid := Spawn(Gen(&TestGenServer{&wg}))
	Cast(pid, TestMessage{1})
	wg.Wait()
	pid.Stop()
}

func TestCall(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	pid := Spawn(Gen(&TestGenServer{&wg}))
	resp := Call(pid, TestMessage{1})
	if resp.(Message).Payload.(TestMessage).id != 1 {
		t.Error("Got wrong response")
	}
	wg.Wait()
	pid.Stop()
}

func BenchmarkCast(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	pid := Spawn(Gen(&TestGenServer{&wg}))
	for i := 0; i < b.N; i++ {
		Cast(pid, TestMessage{1})
	}
	wg.Wait()
	pid.Stop()
}

func BenchmarkCall(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	pid := Spawn(Gen(&TestGenServer{&wg}))
	for i := 0; i < b.N; i++ {
		resp := Call(pid, TestMessage{i})
		if resp.(Message).Payload.(TestMessage).id != i {
			b.Error("Got wrong response")
		}
	}
	wg.Wait()
	pid.Stop()
}