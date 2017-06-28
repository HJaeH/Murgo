package servermodule

import (
	"errors"
	"murgo/pkg/servermodule/log"
	"reflect"
)

const (
	defaultBuf = 256
)

type Module interface {
	Init()
}

type module struct {
	mid string
	sup *modManager
	val reflect.Value

	events    map[int]*Event
	semaphore chan bool
	buf       chan bool
}

type Event struct {
	module *module
	name   string
	val    reflect.Value
	key    int
}

type AsyncCallMessage struct {
	eventVal reflect.Value
	args     []interface{}
	eventKey int

	semaphore chan bool
	buf       chan bool
}

type CallMessage struct {
	modId    string
	eventKey int
	args     []interface{}

	semaphore chan bool
	eventVal  reflect.Value
	buf       chan bool
	reply     chan bool
}

func AddModule(serverMod ServerModule, mod Module, semaphore int) error {
	// get module name string
	mid := getMid(mod)
	smid := getMid(serverMod)

	if modManager, ok := dispatcher.modManager[smid]; ok {
		dispatcher.modManager[mid] = modManager
		gen := newModule(serverMod, mod, semaphore)
		modManager.addChild(mid, gen)
	} else {
		log.Errorf("Error adding module %s", mid)
		return errors.New("Error adding module")
	}
	mod.Init()
	return nil

}

func newModule(smod ServerModule, mod Module, semaphore int) *module {
	newMod := new(module)
	newMod.init(smod, mod, semaphore)
	return newMod
}
func newEvent(mod *module, val reflect.Value, eventKey int) *Event {
	newEvent := new(Event)
	newEvent.module = mod
	newEvent.val = val
	newEvent.key = eventKey
	return newEvent
}

func EventRegister(eventInter interface{}, eventKey int) {
	modName, eventName := rawReqParser(eventInter)
	if err := register(modName, eventName, eventKey); err != nil {
		log.Errorf("Failed to register event at %s, event: %s", modName, eventName)
		return
	}
}

func register(modName string, eventName string, eventKey int) error {

	sup, _ := dispatcher.modManager[modName]
	mod, err := sup.child(modName)
	if err != nil {
		log.Errorf("Module does not exist : %s", modName)
		return err
	}
	eventVal := mod.val.MethodByName(eventName)
	newEvent := newEvent(mod, eventVal, eventKey)

	mod.events[eventKey] = newEvent
	dispatcher.eventMap[eventKey] = newEvent
	return nil
}

func getMid(mod interface{}) string {
	return reflect.TypeOf(mod).Elem().Name()
}

func (m *module) init(modManager ServerModule, mod Module, semaphore int) {
	mid := getMid(mod)
	mmid := getMid(modManager)

	parent := dispatcher.modManager[mmid]
	m.buf = make(chan bool, defaultBuf)
	m.semaphore = make(chan bool, semaphore)
	m.events = make(map[int]*Event)
	m.mid = mid
	m.sup = parent
	m.val = reflect.ValueOf(mod)
}
