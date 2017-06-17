package servermodule

import (
	"fmt"
	"sync"
)

const (
	routerQueueSize = 5
)

type Router struct {
	callRouter      chan *CallMessage
	asyncCallRouter chan *AsyncCallMessage

	modManager map[string]*modManager
	apiMap     map[int]*API
}

//global singleton router
var router *Router
var once sync.Once

func newRouter(modId string, modManager *modManager) *Router {

	once.Do(func() {
		router = &Router{}
	})
	router.init()
	router.modManager[modId] = modManager
	go router.run()
	return router
}

func (r *Router) init() {
	r.asyncCallRouter = make(chan *AsyncCallMessage, routerQueueSize)
	r.callRouter = make(chan *CallMessage)
	r.modManager = make(map[string]*modManager)
	r.apiMap = make(map[int]*API)
}
func (r *Router) run() {
	for {
		select {
		//todo : cast, call 반복 코드 제거
		case m := <-r.callRouter:
			//fmt.Println(m.apiKey, "call received")
			api := r.getAPI(m.apiKey)
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
		case m := <-r.asyncCallRouter:
			api := r.getAPI(m.apiKey)
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
				panic("module buffer is full and message is lost, need to change capacity")
			}

		}
	}
}

func asyncCall(key int, args ...interface{}) {
	router.asyncCallRouter <- &AsyncCallMessage{
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

	router.callRouter <- msg
	<-reply
}

func (r *Router) getAPI(apiKey int) *API {
	if api, ok := r.apiMap[apiKey]; ok {
		return api
	} else {
		fmt.Println("Invalid API key: %s", apiKey)
		panic("Invalid API key")
	}

}

func AsyncCall(key int, args ...interface{}) {
	asyncCall(key, args...)

}

//todo : return result - issue : specifying return value type.
func Call(key int, args ...interface{}) {
	call(key, args...)
}
