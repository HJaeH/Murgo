package servermodule

import (
	"errors"
	"murgo/pkg/servermodule/log"
	"sync"
)

type Dispatcher struct {
	call      chan *CallMessage
	asyncCall chan *AsyncCallMessage

	modManager map[string]*modManager
	eventMap   map[int]*Event
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
	d.eventMap = make(map[int]*Event)
}

func (d *Dispatcher) run() {
	defer restart(d.run)
	d.dispatchLoop()
}
func (d *Dispatcher) dispatchLoop() {
	for {
		select {
		//todo : cast, call 반복 코드 제거
		case m := <-d.call:
			if event, err := d.getEvent(m.eventKey); err != nil {
				log.Error(err, "Event key is not exist")
			} else {
				mod := event.module
				select {
				case mod.buf <- true:
					mod.sup.callChan <- &CallMessage{
						args:      m.args,
						eventKey:  m.eventKey,
						reply:     m.reply,
						eventVal:  event.val,
						semaphore: mod.semaphore,
						buf:       mod.buf,
					}
				}
			}

		case m := <-d.asyncCall:
			if event, err := d.getEvent(m.eventKey); err != nil {
				log.Error(err, "Event key is not exist")
			} else {
				mod := event.module
				select {
				case mod.buf <- true:
					mod.sup.asyncCallChan <- &AsyncCallMessage{
						args:      m.args,
						eventKey:  m.eventKey,
						eventVal:  event.val,
						semaphore: mod.semaphore,
						buf:       mod.buf,
					}
				default:
					log.Warnf("Message lost at module : %s", mod.mid)
				}
			}
		}
	}
}

func asyncCall(eventKey int, args ...interface{}) {
	dispatcher.asyncCall <- &AsyncCallMessage{
		args:     args,
		eventKey: eventKey,
	}
}

func call(eventKey int, args ...interface{}) {
	reply := make(chan bool)
	msg := &CallMessage{
		eventKey: eventKey,
		args:     args,
		reply:    reply,
	}

	dispatcher.call <- msg
	<-reply
}

func (d *Dispatcher) getEvent(eventKey int) (*Event, error) {
	//todo: concurrency test
	//r.mux.Lock()
	//r.mux.Unlock()
	if event, ok := d.eventMap[eventKey]; ok {
		return event, nil
	} else {
		return nil, errors.New("Invalid Event key")
	}

}

func AsyncCall(key int, args ...interface{}) {
	asyncCall(key, args...)

}

//todo : return result - issue : specifying return type.
func Call(key int, args ...interface{}) {
	call(key, args...)
}
