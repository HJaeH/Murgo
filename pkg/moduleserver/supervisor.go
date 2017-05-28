// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor

package moduleserver

import (
	"fmt"
	"reflect"
)

//cast, call 통신을 위해서는 모든 genServer에 접근 할 수 있는 모듈이 필요하고
//실시간으로 각 genServer로 부터 cast, call 을 처리 할 수 있어야 함

type Supervisor struct {
	//supManager
	parent    *Supervisor
	isRunning bool
	children  map[string]GenServer
	castChan  chan *CastMessage
	callChan  chan *CallMessage
}

type SupervisorInt interface {
	init()
	startChild()
}

//todo supmap lock??
var supMap = make(map[string]*Supervisor)
var supMapTemp = make(map[int]*Supervisor)

/*func TerminateChild(supervisor interface{}, pid chan int) {
	Call(supervisor, pid)
	//call(Supervisor, {terminate_child, Name}).
}*/

func Call(moduleName string, msg *CallMessage) {

	supMap[moduleName].callChan <- msg

}
func Cast(moduleName string, msg *CastMessage) {
	supMap[moduleName].castChan <- msg

}

/*func startsupervisor() chan childchan {



	return childchan

}*/
func StartSupervisor(module SupervisorInt) {
	//todo : supervisor pkg instance하나 만들어서 입력받은 string과 매칭 해야함
	/*if ok, _ := supMap[sup]; ok {

	}
	supervisor := new(Supervisor)*/
	//startServerManager()
	supervisor := new(Supervisor)
	module.init()
	supervisor.run()
}

//todo : 마저 구현
func StartLinkSupervisor() {
	supervisor := new(Supervisor)
	supervisor.run()
}

//add child to supervisor
func (supervisor *Supervisor) link(serverName string, module GenServer) {
	supervisor.children[serverName] = module
}

// get child module by name
func (supervisor *Supervisor) child(moduleName string) GenServer {
	return supervisor.children[moduleName]
}

func (supervisor *Supervisor) run() {
	defer restart(supervisor.supervisorLoop)
	supervisor.supervisorLoop()

}

func (supervisor *Supervisor) supervisorLoop() {
	// todo : single goroutine genserver랑 아닌거 구분 해서 구현
	for {
		//todo : goroutine pool 구현
		//req(cast, call)단위로 goroutine작업 수행 함으로서 비동기 동작 구현
		select {
		case msg := <-supervisor.callChan:
			//todo : call return 구현, 동기 코드 구현
			//go supervisor.child(msg.moduleName).handleCast(msg)
			go supervisor.handleCall(msg)
		case msg := <-supervisor.castChan:
			go supervisor.handleCast(msg)
		}
	}
}

func (supervisor *Supervisor) handleCast(msg *CastMessage) {
	module := supervisor.children[msg.moduleName]
	funcName := msg.funcName
	args := msg.args
	inputs := make([]reflect.Value, len(args))

	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])

	}
	reflect.ValueOf(module).MethodByName(funcName).Call(inputs)
}

func (supervisor *Supervisor) handleCall(msg *CallMessage) {
	module := supervisor.children[msg.moduleName]
	funcName := msg.funcName
	args := msg.args
	inputs := make([]reflect.Value, len(args))

	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])

	}
	reflect.ValueOf(module).MethodByName(funcName).Call(inputs)
	/*returnChan <- &CallReply{
		//sender: ,

	}*/ // todo : call return 작업중
}

func restart(module func()) {
	if err := recover(); err != nil {
		fmt.Println("Recovered from panic")
		module()
	}
}
