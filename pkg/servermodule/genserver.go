package servermodule

import (
	"reflect"
	"runtime"
)

type mid string
type api string

//todo : req type 변경시에 타입 지정해서 사용
//type request string

type GenCallback interface {
	Init()
	//Remove()
	HandleCast(string, ...interface{})
	HandleCall(string, ...interface{})
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

func StartLinkGenServer(smod SupCallback, mod GenCallback) {

	// get module name string
	mid := getMid(mod)
	smid := getMid(smod)

	//set supervisor of new module
	if sup, ok := GlobalParent[smid]; ok {
		GlobalParent[mid] = sup
		sup.addChild(mid, mod)
	} else {
		panic("Invalid supName")
	}
	mod.Init()

}

func RegisterAPI(rawAPI interface{}) {
	modId, apiName := rawReqParser(rawAPI)
	rawStr := runtime.FuncForPC(reflect.ValueOf(rawAPI).Pointer()).Name()

	register(modId, apiName)
}

func register(modId mid, apiName api) {

	GlobalParent[modId].apiMap[apiName]
	/*if _, ok := APIMap[apiKey]; !ok {
		APIMap[apiKey] = request
		return
	}*/
	panic("Duplicated API Key")

}

func getMid(mod interface{}) mid {
	//var modName interface{}
	modName := reflect.TypeOf(mod).Elem().Name()
	return mid(modName)
}
