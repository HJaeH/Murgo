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
	cm *ChannelManager
	ts *TlsServer
	tc map[uint32]*TlsClient

	Cast chan interface{}
	Call chan interface{}


}

func NewSupervisor() (*Supervisor){
	supervisor := new(Supervisor)
	supervisor.Cast = make(chan interface{})
	supervisor.tc = make( map[uint32]*TlsClient)


	supervisor.mh = NewMessageHandler(supervisor)
	supervisor.ts = NewTlsServer(supervisor)
	supervisor.cm = NewChannelManager(supervisor)

	return supervisor
}

func (supervisor *Supervisor)StartSupervisor() {

	//go startChannelManager(supervisor) // TODO
	//go startSessionManager(supervisor) // TODO

	go supervisor.mh.startMassageHandler() // rumble_server 같은 역할
	go supervisor.ts.startTlsServer()

	for{
		select {
		//supervisor cast channel로 session값이 오면 newclient
		case castData := <-supervisor.Cast:
			supervisor.handleCast(castData)
		default:
		}
	}
}



func (supervisor *Supervisor) handleCast (castData interface{}) {

	switch t := castData.(type) {
	default:
		fmt.Printf(" : unexpected type %T", t)
	/*case uint32:
		session := castData.(uint32)
		go supervisor.tc[session].startTlsClient()*/
	}

}



func (supervisor *Supervisor)startGenServer(genServer func()) {
	fmt.Println("gen server started")

	defer func(){
		if err:= recover(); err!= nil{
			fmt.Println("Fail to recover client")
		}
	}()
	go genServer()
}
