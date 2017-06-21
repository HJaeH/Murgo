package server

import (
	"fmt"
	"murgo/config"
	"murgo/pkg/servermodule"
)

type ModManager struct{}

func Start() {

	servermodule.Start(new(ModManager)) // root supervisor
}
func Terminate() error {
	return servermodule.Terminate()
}

//callback
func (m *ModManager) Init() {

	//server modules
	servermodule.AddModule(m, new(SessionManager), 1)
	servermodule.AddModule(m, new(ChannelManager), 1)
	servermodule.AddModule(m, new(Server), 100)
	servermodule.AddModule(m, new(MessageHandler), 5)
	fmt.Println(config.AppName, "is now running")
}

func (m *ModManager) Terminate() error {
	return nil
}
