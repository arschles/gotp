package gotp

type GenServer interface {
	Init()
	HandleCall(Message) interface{}
	HandleCast(Message)
}

type GoGenServer struct {
	GoActorWithInit

	server GenServer 
}

func (serv *GoGenServer) Init() {
	serv.server.Init()
}

func (serv *GoGenServer) Receive(msg Message) error {
 	switch m := msg.Payload.(type) {
	case cast:
		serv.server.HandleCast(Message{m.message})
	case call: 
		m.respChan <- serv.server.HandleCall(Message{m.message})
	default:
		serv.server.HandleCast(Message{msg})
	}
	return nil
}

type call struct {
	respChan chan interface{}
	message interface{}
}

type cast struct {
	message interface{}
}

func Gen(gen GenServer) Actor {
	return &GoGenServer{server: gen}
}

func Call(pid Pid, msg interface{}) interface{} {
	respChan := make(chan interface{})
	callMsg := call{respChan, msg}
	pid.Send(callMsg)
	return <- respChan
}

func Cast(pid Pid, msg interface{}) {
	pid.Send(cast{msg})
}