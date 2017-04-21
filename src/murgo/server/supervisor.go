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

<<<<<<< HEAD
	Cast chan interface{}
	Call chan interface{}


}

func NewSupervisor() (*Supervisor){
	supervisor := new(Supervisor)
	supervisor.Cast = make(chan interface{})
=======
	cast chan interface{}

}

//supervisor를 루트로 하는 double linked의 tree형태.
//4개의 고루틴을 실행시킨다
func NewSupervisor() (*Supervisor){
	supervisor := new(Supervisor)
	supervisor.cast = make(chan interface{})
>>>>>>> 85f3bc2cc36f45fdc4e241541f4528d9a68e2290
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

<<<<<<< HEAD
	for{
		select {
		//supervisor cast channel로 session값이 오면 newclient
		case castData := <-supervisor.Cast:
=======
	//var a = 10
	for{
		//supervisor.cast <- a
		select {
		//supervisor cast channel로 session값이 오면 newclient
		case castData := <-supervisor.cast:
>>>>>>> 85f3bc2cc36f45fdc4e241541f4528d9a68e2290
			supervisor.handleCast(castData)
		default:
		}
	}
}



func (supervisor *Supervisor) handleCast (castData interface{}) {

<<<<<<< HEAD
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
=======

	switch t := castData.(type) {
	default:
		fmt.Printf(" : unexpected type %T", t)
	case uint32:
		session := castData.(uint32)
		go supervisor.tc[session].startTlsClient()
	}
>>>>>>> 85f3bc2cc36f45fdc4e241541f4528d9a68e2290
}
