package servermodule


type Module interface {
	startLink()
}



func start(module Module){
	module.startLink()
}


