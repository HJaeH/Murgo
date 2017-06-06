package servermodule

import (
	"fmt"
	"sync"
)

const (
	singleGoRoutineSize   = 1
	multipleGoRoutineSize = 3

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

			select {
			//check whether this module is available now or not
			case r.apiMap[m.apiKey].module.sync <- true:

				api := r.apiMap[m.apiKey]
				m.apiVal = api.val
				//msg.temp.val = api.val

				//msg.syncChan = r.apiMap[msg.apiKey].module.sync

				api.module.sup.castChan <- &CastMessage{
					args:     m.args,
					apiVal:   m.apiVal,
					syncChan: r.apiMap[m.apiKey].module.sync,
				}

			default:
				fmt.Println("!@#")
			}
			/*inputs := make([]reflect.Value, len(m.args))
			for i, _ := range m.args {
				inputs[i] = reflect.ValueOf(m.args[i])
			}
			fmt.Println(inputs)
			//m.val.Call([]reflect.Value{})
			m.apiVal.Call(inputs)*/
		default:
		}

	}

}

func cast(apiKey int, args ...interface{}) {

	/*if args == nil {
		args = nil
	}
	*/
	//todo : by adding message creator, cope with multiple goroutine case
	/*	router.castRouter <- &CastMessage{
		apiVal: router.apiMap[apiKey].val,
		args:   args,
		apiKey: apiKey,
		//syncChan: make(chan int, singleGoRoutineSize),
	}*/
}

func Cast1(key int, args ...interface{}) {
	fmt.Println(key, args, "============")
	//reflect.ValueOf(temp).MethodByName("HandleIncomingClient").Call([]reflect.Value{})
	/*reflect.ValueOf(temp).MethodByName("HandleIncommingClient").Call([]reflect.Value{})*/

	router.castRouter <- &CastMessage{
		apiVal: router.apiMap[key].val,
		args:   args,
		apiKey: key,
	}

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

func Cast(key int, args ...interface{}) {
	fmt.Println("cast!@#!@!!!!!!!!!!!!!!!!!!!")

	/*if args == nil {
		cast(key)
	} else {
		cast(key, args)

	}
	*/
}
