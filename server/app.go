package server

import (
	"errors"
	"murgo/pkg/servermodule"
)

type ModManager struct{}

func Start() {
	servermodule.Start(new(ModManager)) // root supervisor
}

func (m *ModManager) Terminate() error {
	// todo : terminate
	return errors.New("Error")
}

func Terminate() error {
	err := errors.New("terminate")
	return err
}

//callback
func (m *ModManager) Init() {
	//server modules
	servermodule.AddModule(m, new(SessionManager), 1)
	servermodule.AddModule(m, new(ChannelManager), 1)
	servermodule.AddModule(m, new(Server), 100)
	servermodule.AddModule(m, new(MessageHandler), 5)
}
