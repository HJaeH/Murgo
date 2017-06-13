package server

import (
	"errors"
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

//callback
func (ms *MurgoSupervisor) Init() {

	//server modules
	servermodule.StartLinkGenServer(ms, new(SessionManager), true)
	servermodule.StartLinkGenServer(ms, new(ChannelManager), true)
	servermodule.StartLinkGenServer(ms, new(TlsServer), false)
	servermodule.StartLinkGenServer(ms, new(MessageHandler), false)
}
