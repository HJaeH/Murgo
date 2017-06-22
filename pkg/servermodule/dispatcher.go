package servermodule

import (
	"fmt"
	"sync"
)

type Dispatcher struct {
	call      chan *CallMessage
	asyncCall chan *AsyncCallMessage

	modManager map[string]*modManager
	apiMap     map[int]*API
	mux        sync.Mutex
}

//global singleton
var dispatcher *Dispatcher
var once sync.Once

func newDispatcher(modId string, modManager *modManager) *Dispatcher {

	once.Do(func() {
		dispatcher = &Dispatcher{}
	})
	dispatcher.init()
	dispatcher.modManager[modId] = modManager
	go dispatcher.run()
	return dispatcher
}

func (d *Dispatcher) init() {
	d.asyncCall = make(chan *AsyncCallMessage, 10)
	d.call = make(chan *CallMessage)
	d.modManager = make(map[string]*modManager)
	d.apiMap = make(map[int]*API)
}

func (d *Dispatcher) run() {
	for {
		select {
		//todo : cast, call 반복 코드 제거
		case m := <-d.call:
			//fmt.Println(m.apiKey, "call received")
			api := d.getEvent(m.apiKey)
			mod := api.module
			select {
			case mod.buf <- true:
				mod.sup.callChan <- &CallMessage{
					args:     m.args,
					apiKey:   m.apiKey,
					reply:    m.reply,
					apiVal:   api.val,
					syncChan: mod.semaphore,
					buf:      mod.buf,
				}
			}
		case m := <-d.asyncCall:

			api := d.getEvent(m.apiKey)
			mod := api.module
			select {
			case mod.buf <- true:
				mod.sup.asyncCallChan <- &AsyncCallMessage{
					args:     m.args,
					apiKey:   m.apiKey,
					apiVal:   api.val,
					syncChan: mod.semaphore,
					buf:      mod.buf,
				}
			default:
				fmt.Println("message lost at mod :", mod.mid)
				panic("module buffer overflow occured and message is lost, need to change capacity")
			}

		}
	}
}

func asyncCall(key int, args ...interface{}) {
	dispatcher.asyncCall <- &AsyncCallMessage{
		args:   args,
		apiKey: key,
	}
}

func call(apiKey int, args ...interface{}) {
	reply := make(chan bool)
	msg := &CallMessage{
		apiKey: apiKey,
		args:   args,
		reply:  reply,
	}

	dispatcher.call <- msg
	<-reply
}

func (d *Dispatcher) getEvent(apiKey int) *API {
	//r.mux.Lock()
	api, ok := d.apiMap[apiKey]
	//r.mux.Unlock()

	if ok {
		return api
	} else {
		fmt.Println("Invalid API key: %s", apiKey)
		panic("Invalid API key")
	}

}

func AsyncCall(key int, args ...interface{}) {
	asyncCall(key, args...)

}

//todo : return result - issue : specifying return type.
func Call(key int, args ...interface{}) {
	call(key, args...)
}
