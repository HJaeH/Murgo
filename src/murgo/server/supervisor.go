// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor
// 고루틴을 실행시키고 고루틴간의 통신시 슈퍼바이저를 통해서 각자의 채널에 접근할수 있게 한다


package server

import (
	"fmt"
)




type Supervisor struct {
	sm *SessionManager
	mh *MessageHandler
	ch *ChannelManager
	ts *TlsServer
	tc map[uint32]*TlsClient

	cast chan interface{}

}

//supervisor를 루트로 하는 double linked의 tree형태.
//4개의 고루틴을 실행시킨다
func NewSupervisor() (*Supervisor){
	supervisor := new(Supervisor)
	supervisor.cast = make(chan interface{})
	supervisor.tc = make( map[uint32]*TlsClient)


	supervisor.mh = NewMessageHandler(supervisor)
	supervisor.ts = NewTlsServer(supervisor)

	return supervisor
}

func (supervisor *Supervisor)StartSupervisor() {

	//go startChannelManager(supervisor) // TODO
	//go startSessionManager(supervisor) // TODO

	go supervisor.mh.startMassageHandler() // rumble_server 같은 역할
	go supervisor.ts.startTlsServer()

	//var a = 10
	for{
		//supervisor.cast <- a
		select {
		//supervisor cast channel로 session값이 오면 newclient
		case castData := <-supervisor.cast:
			fmt.Println("ddd")
			supervisor.castHandler(castData)
		default:
		}
	}
}



func (supervisor *Supervisor) castHandler (castData interface{}) {

	fmt.Println("create client")

	switch t := castData.(type) {
	default:
		fmt.Printf(" : unexpected type %T", t)
	case uint32:
		fmt.Println("create client")
		session := castData.(uint32)
		go supervisor.tc[session].startTlsClient()
	}
}
