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

	children map[string]*GenServer
}

func (s *Sup) init() {
	s.isRunning = false
	s.children = make(map[string]*GenServer)
	s.castChan = make(chan *CastMessage)
	s.callChan = make(chan *CallMessage)
}

//todo :
/*func StartLinkSupervisor() {
	supervisor := new(Supervisor)
	supervisor.run()
}*/

//add child to supervisor
func (s *Sup) addChild(mid string, genServer *GenServer) {

	s.children[mid] = genServer
}

// get child module by name
func (s *Sup) child(mid string) *GenServer {
	if mod, ok := s.children[mid]; ok {
		return mod
	}
	panic("Mid doesn't exist")
}

func (s *Sup) run() {

	defer restart(s.supervisorLoop)
	s.supervisorLoop()

}

func (s *Sup) supervisorLoop() {
	if s.isRunning {
		panic("supervisor is already runnning")
	}
	s.isRunning = true
	for {
		//todo : goroutine pool로 goroutine 필요할듯
		select {

		case msg := <-s.callChan:
			//todo : call return 구현, 동기 코드 구현
			go s.handleCall(msg)

		case msg := <-s.castChan:
			/*if msg.wg != nil {
				msg.wg.Add(1)
				msg.wg.Wait()
			}
			msg.wg.Add(1)
			msg.wg.*/

			go s.handleCast(msg)
		}
	}
}

func (s *Sup) handleCast(msg *CastMessage) {
	fmt.Println("Cast API", msg.apiKey)

	/*if msg.args == nil {
		msg.apiVal.Call([]reflect.Value{})
	} else {
		doCall(msg.apiVal, msg.args...)
	}*/
	select {
	//check whether this module is available now or not
	case msg.syncChan <- true:
		doCall(msg.apiVal, msg.args...)
	default:
		fmt.Println("buffer is full")
		cast(msg.apiKey, msg.args...)
	}

	defer func() {
		<-msg.syncChan
		<-msg.buf

	}()

}

func doCall(val reflect.Value, args ...interface{}) {

	if args == nil {
		val.Call([]reflect.Value{})
	} else {
		inputs := make([]reflect.Value, len(args))
		for i, _ := range args {
			inputs[i] = reflect.ValueOf(args[i])
		}

		val.Call(inputs)
	}
}

func (s *Sup) handleCall(msg *CallMessage) {
	/*if msg.args == nil {
		msg.apiVal.Call([]reflect.Value{})
	} else {
		inputs := make([]reflect.Value, len(msg.args))
		for i, _ := range msg.args {
			inputs[i] = reflect.ValueOf(msg.args[i])

		}

		msg.apiVal.Call(inputs)
	}*/
	select {
	//check whether this module is available now or not
	case msg.syncChan <- true:
		//execute the request
		doCall(msg.apiVal, msg.args...)
	default:
		//call again
		call(msg.apiKey, msg.args)
	}

	defer func() {
		<-msg.syncChan
		<-msg.buf
		//cast sync
		msg.reply <- true

	}()
}

func restart(module func()) {
	if err := recover(); err != nil {
		fmt.Println("Recovered from panic")
		module()
	}
}

func toSlice(args ...interface{}) []interface{} {
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

func rawReqParser(rawAPI interface{}) (string, string) {

	//returns string ex)main.(*B).A
	rawStr := runtime.FuncForPC(reflect.ValueOf(rawAPI).Pointer()).Name()
	reqs := strings.Split(rawStr, ".")
	modName := strings.Trim(reqs[len(reqs)-2], "(*)")

	apiName := reqs[len(reqs)-1]

	return modName, apiName
}

// todo : terminate구현
/*func TerminateChild(supervisor interface{}, pid chan int) {
	Call(supervisor, pid)
	//call(Supervisor, {terminate_child, Name}).
}*/

//// externals
func StartSupervisor(smod SupCallback) {

	smid := getMid(smod)

	supervisor := new(Sup)
	supervisor.init()

	//run router
	newRouter(smid, supervisor)

	router.supervisors[smid] = supervisor
	//run supervisor
	go supervisor.run()

	//app's Init callback
	smod.Init()
}
