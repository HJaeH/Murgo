package servermodule

import (
	"reflect"
	"sync"
)

const (
	single     = 1
	multi      = 5
	defaultBuf = 10
)

type GenCallback interface {
	Init()
}

type GenServer struct {
	mid string
	sup *Sup
	val reflect.Value

	apis map[int]*API
	sync chan bool
	buf  chan bool

	//todo : 코드 개선
	wg *sync.WaitGroup
}

type API struct {
	module *GenServer
	name   string
	val    reflect.Value
	key    int
}

type CastMessage struct {
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

type CallReply struct {
	result int
}

func StartLinkGenServer(smod SupCallback, mod GenCallback, singleThreaed bool) {
	//genserver에 module을 등록.
	// get module name string

	mid := getMid(mod)
	smid := getMid(smod)

	//set supervisor of new module
	if sup, ok := router.supervisors[smid]; ok {
		router.supervisors[mid] = sup
		gen := newGenServer(smod, mod, singleThreaed)
		sup.addChild(mid, gen)
	} else {
		panic("Invalid supName")
	}
	mod.Init()

}

func newGenServer(smod SupCallback, mod GenCallback, sg bool) *GenServer {
	gen := new(GenServer)
	gen.init(smod, mod, sg)
	return gen
}

func newAPI(mod *GenServer, val reflect.Value, apiKey int) *API {
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

	sup, _ := router.supervisors[modName]
	mod := sup.child(modName)
	apiVal := mod.val.MethodByName(apiName)
	/*if mod.val.IsValid() == true {
		fmt.Println(apiName, " is valid mod", modName)
	}

	if apiVal.IsValid() != true {
		panic("invalid api value")
	}*/
	newAPI := newAPI(mod, apiVal, apiKey)

	mod.apis[apiKey] = newAPI
	router.apiMap[apiKey] = newAPI

	/*if sup, ok := router.supervisors[modName]; !ok {
		mod := sup.child(modName)

		newAPI := new(API)
		newAPI.val = mod.val.MethodByName(apiName)
		mod.apis[apiKey] = newAPI
		router.apiMap[apiKey] = newAPI
		return
	}
	panic("Module is not registered")*/
}

func getMid(mod interface{}) string {
	return reflect.TypeOf(mod).Elem().Name()
}

func (g *GenServer) init(smod SupCallback, mod GenCallback, sg bool) {
	//todo mid 추가 여부
	mid := getMid(mod)
	smid := getMid(smod)

	parent := router.supervisors[smid]
	g.wg = new(sync.WaitGroup)
	//g.sync = make(chan bool, routeBufferPerModule)
	g.buf = make(chan bool, defaultBuf)
	if sg {
		g.sync = make(chan bool, single)
	} else {
		g.sync = make(chan bool, multi)
	}
	g.apis = make(map[int]*API)
	g.mid = mid
	g.sup = parent
	g.val = reflect.ValueOf(mod)
}
