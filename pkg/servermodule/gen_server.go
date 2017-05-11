package servermodule

import (
	"fmt"
)

type GenServer struct {
	Supervisor
	callback
	CastChannel     chan []interface{}
	CallChannel     chan []interface {}

	serverResources interface{}
}

func (genServer *GenServer) start (){

	// todo : server loop start here
	// todo : who call this ?
	genServer.CallChannel = make (chan []interface{})
	genServer.CastChannel = make (chan []interface{})

	genServer.init()


	defer func(){
		if err:= recover(); err!= nil{
			fmt.Println("Channel manager recovered")
			genServer.start()
		}
	}()

	for {
		select {
		case callData := <-genServer.CallChannel:
			_ = callData // todo : for temporary test
		case castData := <-genServer.CastChannel:
			_ = castData
		}
	}


}




func (genServer *GenServer) Init () {
	//genServer
}

func (genServer *GenServer) Cast ( resources interface{} , data interface{}){

}

func (genServer *GenServer) Call (){

}

