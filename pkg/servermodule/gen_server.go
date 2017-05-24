package servermodule

import (
	"fmt"

)
//type Callback interface {
//	handleCast()
//
//}


type genServer struct {

	castChannel  chan *CastMessage
	callChannel  chan *CallMessage
	closeChannel chan int

	isRunning    bool
	isSupervisor bool

	resource     interface{}
	supervisor   *Supervisor

	genCallback
}

type genCallback interface {
	handleCast(msg *CastMessage)
	handleCall(msg *CallMessage)
	handleInfo()
	terminate()
	init()
}

type CastMessage struct {

	moduleName string
	funcName   string
	args       []interface {}
}

type CallMessage struct {
	moduleName string
	funcName string
	args []interface {}
	returnChan chan *CallReply
}

type CallReply struct {
	sender *genServer
	result int
}

/*// todo : pkg독립 후 이름 startlink로 변환
func StartLinkGenserver(module genCallback){
	genServer := new(genServer)
	genServer.callChannel = make(chan *CallMessage)
	genServer.castChannel = make(chan *CastMessage)
	genServer.isRunning = false
	genServer.isSupervisor = false
	//todo : add parent
	//genServer.supervisor = parent
	module.init()
}*/
/*

func (genServer *genServer)StartLink(parent *Supervisor, module genCallback) {
	genServer.callChannel = make(chan *CallMessage)
	genServer.castChannel = make(chan *CastMessage)
	genServer.isRunning = false
	genServer.isSupervisor = false
	genServer.supervisor = parent
	module.init()
}
*/



/*func (genServer *GenServer) init() {

}*/
func Cast(target string, msg *CastMessage) {

	genServer.supervisor.sendCast(target, msg)
}
func Call (target string, msg *CallMessage){
	genServer.supervisor.sendCall(target, msg)
}



func (genserver *genServer) Stop(){

}
//////////////////////////
/// internal functions ///
//////////////////////////
func (genServer *genServer) start(module genCallback){

	module.init()
	if genServer.isRunning {
		fmt.Println("Genserver is already running on server")
		return
	}
	genServer.isRunning = true

	defer func(){
		if err:= recover(); err!= nil{
			fmt.Println("genserver recovered")
			genServer.start(module)
		}
	}()

	for {
		select {
		case callData := <-genServer.callChannel:
			_ = callData
			//genServer.handleCall(callData)
		case castData := <-genServer.castChannel:
			_ = castData
			//genServer.handleCast(castData)
		case <-genServer.closeChannel:
			return // todo: 그냥 return은 recover 안걸리는지 체크
				// todo: memory free garbage collection 체크
		}
	}

}

/*
func (genServer *GenServer) handleCast(castData interface{}){
	temp := castData.(CastMessage)
	funcName := temp.funcName
	args := temp.args

	inputs := make([]reflect.Value, len(args))

	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])

	}
	reflect.ValueOf(genServer).MethodByName(funcName).Call(inputs)
}

func (genServer *GenServer) handleCall(msg *CallMessage){
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

*/
