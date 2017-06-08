package servermodule

import (
	"fmt"
	"sync"
)

const (
	routerQueueSize = 5
)

type Router struct {
	callRouter chan *CallMessage
	castRouter chan *CastMessage

	supervisors map[string]*Sup
	apiMap      map[int]*API
}

//global singleton router
var router *Router
var once sync.Once

func newRouter(modId string, sup *Sup) *Router {

	once.Do(func() {
		router = &Router{}
	})
	router.init()

	//todo : access supervisor by map or slice
	// set a root supervisor
	router.supervisors[modId] = sup
	go router.run()
	return router
}

func (r *Router) init() {
	r.castRouter = make(chan *CastMessage, routerQueueSize)
	r.callRouter = make(chan *CallMessage)
	r.supervisors = make(map[string]*Sup)
	r.apiMap = make(map[int]*API)
}
func (r *Router) run() {
	for {
		select {
		//todo : cast, call 반복 코드 제거
		case m := <-r.callRouter:
			fmt.Println(m.apiKey, "call received")
			api := r.getAPI(m.apiKey)
			mod := api.module
			select {
			case mod.buf <- true:
				fmt.Println("In ", mod.mid, ", the num of running gorouines : ", len(mod.buf))
				/*api := r.getAPI(m.apiKey)
				mod := api.module*/
				mod.sup.callChan <- &CallMessage{
					args:     m.args,
					apiKey:   m.apiKey,
					reply:    m.reply,
					apiVal:   api.val,
					syncChan: mod.sync,
					buf:      mod.buf,
				}
			}
		case m := <-r.castRouter:
			fmt.Println(m.apiKey, "cast received")
			api := r.getAPI(m.apiKey)
			mod := api.module
			select {
			case mod.buf <- true:
				fmt.Println("In ", mod.mid, ", the num of running gorouines : ", len(mod.buf))
				/*api := r.getAPI(m.apiKey)
				mod := api.module*/
				mod.sup.castChan <- &CastMessage{
					args:     m.args,
					apiKey:   m.apiKey,
					apiVal:   api.val,
					syncChan: mod.sync,
					buf:      mod.buf,
				}
			default:
				fmt.Println("message lost at mod :", mod.mid)
				panic("module buffer is full and message is lost, need to change capacity")
			}

		}
	}
}

func cast(key int, args ...interface{}) {
	router.castRouter <- &CastMessage{
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

func Cast(key int, args ...interface{}) {
	cast(key, args...)

}

func Call(key int, args ...interface{}) {
	call(key, args...)
}
