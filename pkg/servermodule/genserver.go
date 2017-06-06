package servermodule

import (
	"reflect"
)

//todo : req type 변경시에 타입 지정해서 사용
//type request string

type GenCallback interface {
	Init()
	//Remove()
	//HandleCast(string, ...interface{})
	//HandleCall(string, ...interface{})
}

type GenServer struct {
	sup *Sup
	val reflect.Value

	apis map[int]*API
	sync chan bool
}

type API struct {
	module *GenServer
	name   string
	val    reflect.Value
	key    int
}

/*
type CastMessage struct {
	//modId  string
	apiKey int

	args []interface{}

	//by supervisor to call api
	//syncChan chan bool
	apiVal reflect.Value
}*/
type CastMessage struct {
	apiVal reflect.Value
	args   []interface{}
	apiKey int

	//by supervisor to call api
	syncChan chan bool
}

type CallMessage struct {
	modId  string
	apiKey int
	args   []interface{}

	//by supervisor to call api
	syncChan chan bool
	apiVal   reflect.Value

	reply chan *CallReply
}

type CallReply struct {
	result int
}

func StartLinkGenServer(smod SupCallback, mod GenCallback) {
	//genserver에 module을 등록.
	// get module name string

	mid := getMid(mod)
	smid := getMid(smod)

	//set supervisor of new module
	if sup, ok := router.supervisors[smid]; ok {
		router.supervisors[mid] = sup
		gen := newGenServer(smod, mod)
		sup.addChild(mid, gen)
	} else {
		panic("Invalid supName")
	}
	mod.Init()

}

func newGenServer(smod SupCallback, mod GenCallback) *GenServer {
	gen := new(GenServer)
	gen.init(smod, mod)
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

func (g *GenServer) init(smod SupCallback, mod GenCallback) {
	//todo mid 추가 여부
	//mid := getMid(mod)
	smid := getMid(smod)
	parent := router.supervisors[smid]

	g.sync = make(chan bool, 1)
	g.apis = make(map[int]*API)

	g.sup = parent
	g.val = reflect.ValueOf(mod)
}
