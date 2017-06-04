package servermodule

import (
	"reflect"
)

type mid int

//type api string

type apiKey int
type api reflect.Value

//todo : req type 변경시에 타입 지정해서 사용
//type request string

type GenCallback interface {
	Init()
	//Remove()
	HandleCast(string, ...interface{})
	HandleCall(string, ...interface{})
}

type GenServer struct {
	parent *Sup
	router map[apiKey]reflect.Value
	module interface{}
	val    reflect.Value
	name   string
}

type API struct {
	module interface{}
	name   string
	val    reflect.Value
}

type CastMessage struct {
	modId   mid
	apiName api
	args    []interface{}
}

type CallMessage struct {
	modId   mid
	apiName api
	args    []interface{}
	reply   chan *CallReply
}

type CallReply struct {
	result int
}

func StartLinkGenServer(smod mid, mod GenCallback) {
	//genserver에 module을 등록.
	// get module name string
	mid := getMid(mod)
	smid := getMid(smod)

	//set supervisor of new module
	if sup, ok := GlobalParent[smid]; ok {
		GlobalParent[mid] = sup
		gen := newGenServer(mid, smid)
		sup.addChild(mid, gen)
	} else {
		panic("Invalid supName")
	}
	mod.Init()

}

func newGenServer(smod SupCallback, mod GenCallback) GenServer {
	gen := new(GenServer)
	gen.init(smod, mod)
	return gen
}

func RegisterAPI(rawAPI interface{}) {
	modId, apiName := rawReqParser(rawAPI)
	register(modId, apiName)

}

func register(modId mid, apiName api) {

	if sup, ok := GlobalParent[modId]; !ok {
		mod := sup.child(modId)
		newAPI := new(API)
		newAPI.val = mod.val.MethodByName(apiName)
		mod.router[apiName] = newAPI
		return
	}
	panic("Module is not registered")
}

func getMid(mod interface{}) mid {
	//var modName interface{}
	modName := reflect.TypeOf(mod).Elem().Name()
	return mid(modName)
}

func (g *GenServer) init(smod SupCallback, mod GenCallback) {
	//todo mid 추가 여부
	//mid := getMid(mod)
	smid := getMid(smod)
	parent := GlobalParent[smid]

	g.router = make(map[apiKey]api)
	g.parent = parent
	g.val = reflect.ValueOf(mod)

}
