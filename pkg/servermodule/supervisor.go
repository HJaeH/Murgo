// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor

package servermodule

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

//cast, call 통신을 위해서는 모든 genServer에 접근 할 수 있는 모듈이 필요하고
//실시간으로 각 genServer로 부터 cast, call 을 처리 할 수 있어야 함

type SupCallback interface {
	Init()
}
type Sup struct {
	isRunning bool

	castChan chan *CastMessage
	callChan chan *CallMessage

	children map[mid]GenCallback
	apiMap   map[api]string
}

func (s *Sup) init() {
	s.isRunning = false
	s.children = make(map[mid]GenCallback)
	s.castChan = make(chan *CastMessage)
	s.callChan = make(chan *CallMessage)
}

//todo : thread safety check
//todo : supervisor 전체를 관리하는 모듈 추후 필요 현재는 string -> *Sup의 global map 사용
//Root supervisor의 parent는 자신으로 설정
var GlobalParent = make(map[mid]*Sup)

//todo : 마저 구현
/*func StartLinkSupervisor() {
	supervisor := new(Supervisor)
	supervisor.run()
}*/

//add child to supervisor
func (s *Sup) addChild(m mid, module GenCallback) {
	s.children[m] = module
}

// get child module by name
func (s *Sup) child(mid mid) GenCallback {
	return s.children[mid]
}

func (s *Sup) run() {
	defer restart(s.supervisorLoop)
	s.supervisorLoop()

}

func (this *Sup) supervisorLoop() {
	this.isRunning = true
	// todo : single goroutine genserver랑 아닌거 구분 해서 구현
	for {
		//todo : goroutine pool 구현
		//req(cast, call)단위로 goroutine작업 수행 함으로서 비동기 동작 구현
		select {
		case msg := <-this.callChan:
			//todo : call return 구현, 동기 코드 구현
			//go supervisor.child(msg.moduleName).handleCast(msg)
			go this.handleCall(msg)
		case msg := <-this.castChan:
			go this.handleCast(msg)
		}
	}
}

func (s *Sup) handleCast(msg *CastMessage) {

	module := s.children[msg.modId]

	request := s.apiMap[msg.apiName]
	module.HandleCall(request, msg.args)
}

func (s *Sup) handleCall(msg *CallMessage) {

	module := s.children[msg.modId]

	request := s.apiMap[msg.apiName]
	module.HandleCall(request, msg.args)
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

func toArgumentList(args ...interface{}) []interface{} {
	if args != nil {
		newArgs := make([]interface{}, len(args))
		for i, arg := range args {
			newArgs[i] = arg
		}
		if !(len(args) > 1) {
			fmt.Println(newArgs)
		}

		return newArgs
	}

	return nil
}

func rawReqParser(rawAPI interface{}) (mid, api) {
	var modId, apiName interface{}
	//returns string ex)main.(*B).A
	rawStr := runtime.FuncForPC(reflect.ValueOf(rawAPI).Pointer()).Name()
	reqs := strings.Split(rawStr, ".")
	modId = strings.Trim(reqs[len(reqs)-2], "(*)")

	apiName = reqs[len(reqs)-1]

	return modId.(mid), apiName.(api)
}

func cast(rawReq interface{}, args ...interface{}) {
	mid, request := rawReqParser(rawReq)
	msg := &CastMessage{
		modId:   mid,
		apiName: request,
		args:    toArgumentList(args),
	}
	GlobalParent[mid].castChan <- msg
}

func call(rawReq interface{}, args ...interface{}) {
	mid, request := rawReqParser(rawReq)
	msg := &CallMessage{
		modId:   mid,
		apiName: request,
		args:    toArgumentList(args),
	}
	GlobalParent[mid].castChan <- msg
}

// todo : 추후 구현 필요
/*func TerminateChild(supervisor interface{}, pid chan int) {
	Call(supervisor, pid)
	//call(Supervisor, {terminate_child, Name}).
}*/

//// externals

func StartSupervisor(smod SupCallback) {
	mid := getMid(smod)

	supervisor := new(Sup)
	supervisor.init()

	GlobalParent[mid] = supervisor

	go supervisor.run()
	smod.Init()
}

func Call(module mid, api api, args ...interface{}) {
	call(module, api, args)
	// todo : return
	return
}

func Cast(rawReq interface{}, args ...interface{}) {
	cast(rawReq, args)
}
