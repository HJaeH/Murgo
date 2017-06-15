package servermodule

import (
	"reflect"
	"sync"
)

const (
	defaultBuf = 100
)

type ModCallback interface {
	Init()
}

type Module struct {
	mid string
	sup *modManager
	val reflect.Value

	apis      map[int]*API
	semaphore chan bool
	buf       chan bool

	//todo : 코드 개선
	wg *sync.WaitGroup
}

type API struct {
	module *Module
	name   string
	val    reflect.Value
	key    int
}

type AsyncCallMessage struct {
	apiVal reflect.Value
	args   []interface{}
	apiKey int

	//by supervisor to call api
	syncChan chan bool
	wg       *sync.WaitGroup
	buf      chan bool
}

type CallMessage struct {
	modId  string
	apiKey int
	args   []interface{}

	//by supervisor to call api
	syncChan chan bool
	apiVal   reflect.Value
	buf      chan bool
	reply    chan bool
}

func AddModule(smod ModManagerCallback, mod ModCallback, semaphore int) {
	//genserver에 module을 등록.
	// get module name string

	mid := getMid(mod)
	smid := getMid(smod)

	//set supervisor of new module
	if sup, ok := router.modManager[smid]; ok {
		router.modManager[mid] = sup
		gen := newModule(smod, mod, semaphore)
		sup.addChild(mid, gen)
	} else {
		panic("Invalid supName")
	}
	mod.Init()

}

func newModule(smod ModManagerCallback, mod ModCallback, semaphore int) *Module {
	gen := new(Module)
	gen.init(smod, mod, semaphore)
	return gen
}

func newAPI(mod *Module, val reflect.Value, apiKey int) *API {
	newAPI := new(API)
	newAPI.module = mod
	newAPI.val = val
	newAPI.key = apiKey

	return newAPI

}
func RegisterAPI(rawAPI interface{}, apiKey int) {
	modName, apiName := rawReqParser(rawAPI)
	//fmt.Println(modName, apiName, apiKey, "-------")
	register(modName, apiName, apiKey)

}

func register(modName string, apiName string, apiKey int) {

	sup, _ := router.modManager[modName]
	mod := sup.child(modName)
	apiVal := mod.val.MethodByName(apiName)
	newAPI := newAPI(mod, apiVal, apiKey)

	mod.apis[apiKey] = newAPI
	router.apiMap[apiKey] = newAPI
}

func getMid(mod interface{}) string {
	return reflect.TypeOf(mod).Elem().Name()
}

func (g *Module) init(modManager ModManagerCallback, mod ModCallback, semaphore int) {
	mid := getMid(mod)
	mmid := getMid(modManager)

	parent := router.modManager[mmid]
	g.wg = new(sync.WaitGroup)
	//g.sync = make(chan bool, routeBufferPerModule)
	g.buf = make(chan bool, defaultBuf)
	g.semaphore = make(chan bool, semaphore)
	g.apis = make(map[int]*API)
	g.mid = mid
	g.sup = parent
	g.val = reflect.ValueOf(mod)
}
