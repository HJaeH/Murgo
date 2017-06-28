package servermodule

import (
	"errors"
	"murgo/pkg/servermodule/log"
	"reflect"
	"runtime"
	"strings"
)

type ServerModule interface {
	Init()
}
type modManager struct {
	isRunning bool
	children  map[string]*module

	asyncCallChan chan *AsyncCallMessage
	callChan      chan *CallMessage
	ServerModule
}

func (m *modManager) init() {
	m.isRunning = false
	m.children = make(map[string]*module)
	m.asyncCallChan = make(chan *AsyncCallMessage)
	m.callChan = make(chan *CallMessage)
}

func (m *modManager) addChild(mid string, genServer *module) {
	m.children[mid] = genServer
}

func (m *modManager) child(mid string) (*module, error) {
	if mod, ok := m.children[mid]; ok {
		return mod, nil
	} else {
		return nil, errors.New("Mid doesn't exist")
	}

}

func (m *modManager) run() {
	defer restart(m.run)
	m.handleCallLoop()

}

func (m *modManager) handleCallLoop() {

	if m.isRunning {
		log.Error("modManager is already runnning")
		return
	} else {
		m.isRunning = true
		for {
			select {
			case msg := <-m.callChan:
				go m.handleCall(msg)
			case msg := <-m.asyncCallChan:
				go m.handleCallAsync(msg)
			}
		}
	}
}

func (m *modManager) handleCallAsync(msg *AsyncCallMessage) {
	select {
	//check whether this module is available now or not
	case msg.semaphore <- true:
		doCall(msg.eventVal, msg.args...)
	default:
		log.Error("buffer is full in", msg.eventKey)
		asyncCall(msg.eventKey, msg.args...)
	}
	defer func() {
		<-msg.semaphore
		<-msg.buf
	}()

}

func (m *modManager) handleCall(msg *CallMessage) {
	select {
	//check whether this module is available now or not
	case msg.semaphore <- true:
		//execute the request
		doCall(msg.eventVal, msg.args...)
	default:
		//call again
		call(msg.eventKey, msg.args)
	}

	defer func() {
		<-msg.semaphore
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

func restart(rerun func()) {
	if err := recover(); err != nil {

	}
	rerun()
}

func rawReqParser(eventInter interface{}) (string, string) {

	//returns string like "main.(*B).A"
	rawStr := runtime.FuncForPC(reflect.ValueOf(eventInter).Pointer()).Name()
	eventParts := strings.Split(rawStr, ".")
	modName := strings.Trim(eventParts[len(eventParts)-2], "(*)")
	eventName := eventParts[len(eventParts)-1]
	return modName, eventName
}

func Start(serverModule ServerModule) error {
	smid := getMid(serverModule)
	modManager := new(modManager)
	modManager.init()

	if _, err := log.GetLogger(&log.Log{
		Path:      "logs/murgo.log",
		Level:     "debug",
		Formatter: "text",
		Console:   false,
	}); err != nil {
		return err
	}
	//run
	newDispatcher(smid, modManager)

	//set module manager to dispatcher
	dispatcher.modManager[smid] = modManager

	go modManager.run()
	//app's Init callback
	serverModule.Init()

	return nil
}

//todo stop all routines
func Terminate() error {
	return nil
}
