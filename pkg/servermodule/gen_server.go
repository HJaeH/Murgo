package servermodule

import (
	"fmt"
	"reflect"
)

type GenServer struct {
	castChannel  chan *CastMessage
	callChannel  chan *CallMessage
	closeChannel chan int

	isRunning    bool
	isSupervisor bool

	resource interface{}
	callback Callback

	parent *Supervisor
}
type Callback interface {
	Terminate() // todo : interface 함수 상속 관계에서 구현시 어떤게 call 되는지 확인 필요
	Init()
}
type CastMessage struct {
	funcName string
	args     []interface{}
}

type CallMessage struct {
	funcName   string
	args       []interface{}
	returnChan chan *CallReply
}

type CallReply struct {
	sender *GenServer
	result int
}

func NewGenServer(interface{}) *GenServer {
	genServer := new(GenServer)
	genServer.init()
	return genServer
}

func (genServer *GenServer) init() {
	genServer.callChannel = make(chan *CallMessage)
	genServer.castChannel = make(chan *CastMessage)
	genServer.isRunning = false
	genServer.isSupervisor = false
}
func (genServer *GenServer) Cast(target *GenServer, msg *CastMessage) {
	genServer.parent.sendCast(target, msg)

}
func (genServer *GenServer) Call(target *GenServer, msg *CallMessage) {
	genServer.parent.sendCall(target, msg)
}

func (genServer *GenServer) Terminate() {
	//todo 각각 모듈에서 필요한 terminate 작업 필요 -> interface 로 callback 구현
	genServer.closeChannel <- 1
}

//////////////////////////
/// internal functions ///
//////////////////////////
func (genServer *GenServer) start() {
	if genServer.isRunning {
		fmt.Println("Genserver is already running on server")
		return
	}
	genServer.isRunning = true

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("genserver recovered")
			genServer.start()
		}
	}()

	for {
		select {
		case callData := <-genServer.callChannel:
			genServer.handleCall(callData)
		case castData := <-genServer.castChannel:
			genServer.handleCast(castData)
		case <-genServer.closeChannel:
			return // todo: 그냥 return은 recover 안걸리는지 체크
			// todo: memory free garbage collection 체크
		}
	}

}

func (genServer *GenServer) handleCast(castData interface{}) {
	temp := castData.(CastMessage)
	funcName := temp.funcName
	args := temp.args

	inputs := make([]reflect.Value, len(args))

	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])

	}
	reflect.ValueOf(genServer).MethodByName(funcName).Call(inputs)
}

func (genServer *GenServer) handleCall(msg *CallMessage) {
	funcName := msg.funcName
	args := msg.args
	returnChan := msg.returnChan

	inputs := make([]reflect.Value, len(args))

	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])

	}
	reflect.ValueOf(genServer).MethodByName(funcName).Call(inputs)
	returnChan <- &CallReply{
	//sender: ,

	} // todo : call return 작업중

}
