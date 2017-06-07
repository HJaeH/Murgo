package servermodule

import (
	"fmt"
	"sync"
)

const (
	routerQueueSize = 5
)

type Router struct {
	//castRouter  chan *CastMessage
	callRouter  chan *CallMessage
	supervisors map[string]*Sup
	apiMap      map[int]*API
	castRouter  chan *CastMessage
}

//global singleton router
var router *Router
var once sync.Once

//todo : move to router's member
//var apiMap = make(map[apiKey]*API)

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
	//r.castRouter = make(chan *CastMessage, routerQueueSize)
	r.callRouter = make(chan *CallMessage, routerQueueSize)
	r.supervisors = make(map[string]*Sup)
	r.apiMap = make(map[int]*API)

	r.castRouter = make(chan *CastMessage)
}
func (r *Router) run() {
	for {
		select {

		case m := <-r.castRouter:
			//fmt.Println(m.apiKey, " from cast router")
			select {
			//check whether this module is available now or not
			case getAPI(m.apiKey).module.sync <- true:
				api := getAPI(m.apiKey)
				mod := r.apiMap[m.apiKey].module
				api.module.sup.castChan <- &CastMessage{
					args:     m.args,
					apiVal:   api.val,
					syncChan: mod.sync,
					apiKey:   m.apiKey,
					wg:       mod.wg,
				}

			default:
				//todo : 일단 임시방편, module 버퍼가 가득 찬 경우에 메시지 손실 생기는거 방지,
				r.castRouter <- m

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

func Cast(key int, args ...interface{}) {
	cast(key, args...)

}

func call(apiKey int, args ...interface{}) {
	msg := &CallMessage{
		apiKey: apiKey,
		args:   toSlice(args),
		//syncChan: make(chan int, singleGoRoutineSize),
	}

	router.callRouter <- msg
}

func Call(key int, args ...interface{}) bool {
	call(key, args)
	// todo : return value
	return true
}

func getAPI(apiKey int) *API {
	if api, ok := router.apiMap[apiKey]; ok {
		return api
	} else {
		fmt.Println("Invalid API key: %s", apiKey)
		panic("Invalid API key")
	}

}
