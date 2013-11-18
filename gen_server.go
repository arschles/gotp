package gotp

type GenServer interface {
	HandleCall(Message) interface{}
	HandleCast(Message)
	Init()
}

type GoGenServer struct {}

func (serv *GoGenServer) init {}

type GoGenCallServer struct {
  GoGenServer
}

func (serv *GoGenCallServer) HandleCast(msg Message) {
	//no-op
}


//an example call-only gen server
type CallOnlyServer struct {
	//we use ("extend") GoGenCallServer because it implements a default cast method
	GoGenCallServer
}

func (serv *CallOnlyServer) HandleCall(msg Message) interface{} {
	return nil
}