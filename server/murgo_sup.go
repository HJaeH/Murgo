// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor

package server

import (
	"errors"
	"murgo/pkg/servermodule"
)

//todo 이런식으로 선언 및 관리 구상
//const a = servermodule.ModuleRegister()

const (
	//supervisors
	murgosupervisor = (iota + 1) * 100

	//genservers
	sessionmanager
	channelmanager
	messagehandler
	tlsserver
)

type MurgoSupervisor struct {
}

func (murgoSupervisor *MurgoSupervisor) startTlsClient(chan int) {

}

func StartSupervisor() {
	servermodule.StartSupervisor(new(MurgoSupervisor))
}

func (supervisor *MurgoSupervisor) Terminate() error {
	// todo : terminate
	return errors.New("Error")
}

func Terminate() error {
	err := errors.New("d")
	return err
}

//callbacks
func (ms *MurgoSupervisor) Init() {
	servermodule.StartLinkGenServer(ms, new(SessionManager))
	servermodule.StartLinkGenServer(ms, new(ChannelManager))
	servermodule.StartLinkGenServer(ms, new(TlsServer))

}

//todo : startchild 구현, tlsclient -> supervisor로 구현
//todo : startChild -> interface callback
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
