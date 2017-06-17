package servermodule

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type ModManagerCallback interface {
	Init()
}
type modManager struct {
	isRunning bool
	children  map[string]*Module

	asyncCallChan chan *AsyncCallMessage
	callChan      chan *CallMessage
}

func (m *modManager) init() {
	m.isRunning = false
	m.children = make(map[string]*Module)
	m.asyncCallChan = make(chan *AsyncCallMessage)
	m.callChan = make(chan *CallMessage)
}

func (m *modManager) addChild(mid string, genServer *Module) {

	m.children[mid] = genServer
}

// get child module by name
func (m *modManager) child(mid string) *Module {
	if mod, ok := m.children[mid]; ok {
		return mod
	}
	panic("Mid doesn't exist")
}

func (m *modManager) run() {

	defer restart(m.handleCallLoop)
	m.handleCallLoop()

}

func (s *modManager) handleCallLoop() {
	if s.isRunning {

		panic("modManager is already runnning")
	}
	s.isRunning = true
	for {
		select {

		case msg := <-s.callChan:
			go s.handleCall(msg)

		case msg := <-s.asyncCallChan:

			go s.handleCallAsync(msg)
		}
	}
}

func (m *modManager) handleCallAsync(msg *AsyncCallMessage) {

	select {
	//check whether this module is available now or not
	case msg.syncChan <- true:
		doCall(msg.apiVal, msg.args...)
	default:
		fmt.Println("buffer is full in", msg.apiKey)
		asyncCall(msg.apiKey, msg.args...)
	}

	defer func() {
		<-msg.syncChan
		<-msg.buf

	}()

}

func (m *modManager) handleCall(msg *CallMessage) {
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

func doCall(val reflect.Value, args ...interface{}) []reflect.Value {

	if args == nil {
		return val.Call([]reflect.Value{})
	} else {
		inputs := make([]reflect.Value, len(args))
		for i, _ := range args {
			inputs[i] = reflect.ValueOf(args[i])
		}

		return val.Call(inputs)
	}
}

func restart(module func()) {
	if err := recover(); err != nil {
		fmt.Println("Recovered from panic")
		module()
	}
}

func rawReqParser(rawAPI interface{}) (string, string) {

	//returns string like "main.(*B).A"
	rawStr := runtime.FuncForPC(reflect.ValueOf(rawAPI).Pointer()).Name()
	reqs := strings.Split(rawStr, ".")
	modName := strings.Trim(reqs[len(reqs)-2], "(*)")
	apiName := reqs[len(reqs)-1]
	return modName, apiName
}

//// externals
func Start(modManagerInterface ModManagerCallback) {
	mmid := getMid(modManagerInterface)

	modManager := new(modManager)
	modManager.init()

	//run router
	newRouter(mmid, modManager)

	//set module manager to router
	router.modManager[mmid] = modManager

	go modManager.run()
	//app's Init callback
	modManagerInterface.Init()
}
