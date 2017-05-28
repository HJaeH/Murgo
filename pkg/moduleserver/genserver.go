package moduleserver

type GenServer interface {
	init()
	remove()
}

type CastMessage struct {
	moduleName string
	funcName   string
	args       []interface{}
}

type CallMessage struct {
	moduleName string
	funcName   string
	args       []interface{}
	returnChan chan *CallReply
}

type CallReply struct {
	result int
}

func StartLinkGenServer(sid int, moduleName string, module GenServer) {

	module.init() // init genserver
	supMapTemp[sid].link(moduleName, module)

}

/*func StartLink(moduleName string, module GenServer) {

	module.init()
	supervisor.Ca(moduleName, module)
}*/

func Stop() {

}

//////////////////////////
/// internal functions ///
//////////////////////////

func doCall() {

}
