package server

import (
	"errors"
	"fmt"
	"murgo/pkg/servermodule"
)

type MurgoSupervisor struct{}

func Start() {
	servermodule.StartSupervisor(new(MurgoSupervisor)) // root supervisor
}

func (supervisor *MurgoSupervisor) Terminate() error {
	// todo : terminate
	return errors.New("Error")
}

func Terminate() error {
	err := errors.New("terminate")
	return err
}

func (m *MurgoSupervisor) Temp(a int) {
	fmt.Println(a)
}

//callback
func (ms *MurgoSupervisor) Init() {

	//server modules
	servermodule.StartLinkGenServer(ms, new(SessionManager), true)
	servermodule.StartLinkGenServer(ms, new(ChannelManager), true)
	servermodule.StartLinkGenServer(ms, new(TlsServer), false)
	servermodule.StartLinkGenServer(ms, new(MessageHandler), true)
	servermodule.StartLinkGenServer(ms, new(Channel), true)
}

//todo : startchild 구현, tlsclient -> supervisor로 구현
/*
func (ms *MurgoSupervisor) StartChild() {
	servermodule.Call("TlsClient", )
}
*/

/*func (murgoSupervisor *MurgoSupervisor)StartLink(){
	//NewSupervisor calls the init callback
	supervisor := servermodule.StartLinkSupervisor(murgoSupervisor)

	//supervisor.Start(murgoSupervisor)
	murgoSupervisor.supervisor = supervisor
}*/
