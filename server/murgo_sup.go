// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor

package server

import (
	"errors"
	"murgo/pkg/moduleserver"
)

const (
	//supervisors
	murgosupervisor = iota

	//genservers
	sessionmanager
	channelmanager
)

type MurgoSupervisor struct {
	sid int
}

func (murgoSupervisor *MurgoSupervisor) startTlsClient(chan int) {
	//servermodule.

}

func StartSupervisor() {
	murgoSup := new(MurgoSupervisor)
	moduleserver.StartSupervisor(murgoSup)
}

func (supervisor *MurgoSupervisor) Terminate() error {
	// todo : terminate
	return errors.New("Error")
}

//callbacks
func (ms *MurgoSupervisor) init() {

	moduleserver.StartLinkGenServer(murgosupervisor, sessionmanager, new(SessionManager))
	moduleserver.StartLinkGenServer(murgosupervisor, channelmanager, new(ChannelManager))
	//moduleserver.StartLinkGenServer(ms.sid, "sessionManager")
	//moduleserver.StartLinkGenServer(ms.sid, "sessionManager")
}

//todo : startchild 구현, tlsclient -> supervisor로 구현
//todo : startChild -> interface callback
func (ms *MurgoSupervisor) StartChild() {
	moduleserver.Call("TlsClient", &moduleserver.CallMessage{})
}

/*func (murgoSupervisor *MurgoSupervisor)StartLink(){
	//NewSupervisor calls the init callback
	supervisor := servermodule.StartLinkSupervisor(murgoSupervisor)

	//supervisor.Start(murgoSupervisor)
	murgoSupervisor.supervisor = supervisor
}*/
