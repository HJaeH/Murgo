// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor
// 고루틴을 실행시키고 고루틴간의 통신시 슈퍼바이저를 통해서 각자의 채널에 접근할수 있게 한다
package servermodule
import (
//"fmt"
)
import (
	"fmt"
	"reflect"
)

//cast, call 통신을 위해서는 모든 genServer에 접근 할 수 있는 모듈이 필요하고
//실시간으로 각 genServer로 부터 cast, call 을 처리 할 수 있어야 함
type Supervisor struct {
	isRunning bool

	children map[string]genCallback

	castChan chan *CastMessage
	callChan chan *CallMessage
}


type SupCallback interface {
	init()([]interface{})
}

type module interface {
	startLink()
}



func TerminateChild(supervisor interface{}, pid chan int) {
	call (supervisor, pid)
	//call(Supervisor, {terminate_child, Name}).
}

func call(supervisor interface{}, msg *CallMessage){

}

func cast(supervisor interface{}, msg *CastMessage){
	supervisor
}

//Root supervisor 생성
func StartSupervisor(module SupCallback) *Supervisor{
	supervisor := new(Supervisor)
	supervisor.init()

	//callback

	for i, eachModule := range module.init() {
		supervisor.children[i] = eachModule
	}

	go supervisor.run()
	return supervisor
}
func StartLinkGenserver(module genCallback){
	genServer := new(genServer)
	genServer.callChannel = make(chan *CallMessage)
	genServer.castChannel = make(chan *CastMessage)
	genServer.isRunning = false
	genServer.isSupervisor = false
	//todo : add parent
	//genServer.supervisor = parent
	module.init()
}


func (supervisor *Supervisor)Start(module SupCallback){
	go supervisor.run()
}

//Child supervisor 생성
//todo: start link for supervisor
func (supervisor *Supervisor)StartLink(moduleName string, parent *Supervisor) {
	//parent.startChild(serverName, module GenCallback)
	module := parent.child(moduleName)
	parent.link(moduleName, module)
}

func (supervisor *Supervisor)StartChild(moduleName string, module genCallback) chan int {

	temp := make(chan int)// todo : supervisor 에서 channel 관리
	genServer := new(genServer)
	genServer.StartLink(supervisor, module)
	supervisor.link(moduleName, module)

	return temp
}


//add child to supervisor
func (supervisor *Supervisor)link(serverName string, module genCallback) {
	supervisor.children[serverName] = module
}
// get child module by name
func (supervisor *Supervisor)child(moduleName string) genCallback {
	return supervisor.children[moduleName]
}

/*func (supervisor *Supervisor)StartGenServer(genServer *GenServer) {
	fmt.Println("genServer started")
	//set parent
	genServer.supervisor = supervisor
	//Register child genServer if it's not a supervisor
	if !genServer.isSupervisor {
		supervisor.children[genServer.resource] = genServer
	}
	go genServer.start()
}*/

/*func (supervisor *Supervisor)Terminate(){
	for _, eachChild := range supervisor.children {
		eachChild.terminate()
	}
}*/

///////////////////
// internal  //
////////////////////////

func (supervisor *Supervisor)sendCast(target string, msg *CastMessage) {
	//var ok interface{}
	if module, ok := supervisor.children[target]; ok {
		go module.handleCast(msg)
	}
}

func (supervisor *Supervisor)sendCall(target string, msg *CallMessage) chan int {

	if module, ok := supervisor.children[target]; ok {
		module.handleCall(msg)
	} else {

		//todo : panic
	}

	temp := make(chan int) // todo : return channel
	return temp

}


func (supervisor *Supervisor) run(){
	defer restart(supervisor.supervisorLoop)
	supervisor.supervisorLoop()

}

func (supervisor *Supervisor)supervisorLoop(){
	for {
		//cast단위로 goroutine작업 수행 함으로서 비동기 동작 구현
		select {
		case msg := <- supervisor.callChan:
		//todo : call return 구현, 동기 코드 구현
			go supervisor.child(msg.moduleName).handleCall(msg)
		case msg := <- supervisor.castChan:
			go supervisor.child(msg.moduleName).handleCast(msg)
		}
	}
}

func restart( module func()) {
	if err:= recover(); err!= nil {
		fmt.Println("Recovered from panic")
		module()
	}
}

func (supervisor *Supervisor)init(){
	supervisor.children = make(map[string]genCallback)
	supervisor.castChan = make(chan *CastMessage)
	supervisor.callChan = make(chan *CallMessage)
	supervisor.isRunning = false
}