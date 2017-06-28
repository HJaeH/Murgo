package server

import (
	"murgo/config"
	"murgo/pkg/servermodule"
	"murgo/pkg/servermodule/log"
)

type ServerModule struct{}

const SingleThread = 1

func Start() {

	if err := servermodule.Start(new(ServerModule)); err != nil {

	}
}
func Terminate() error {
	return servermodule.Terminate()
}

//callback
func (s *ServerModule) Init() {

	//server modules
	servermodule.AddModule(s, new(SessionManager), SingleThread)
	servermodule.AddModule(s, new(ChannelManager), SingleThread)
	servermodule.AddModule(s, new(Server), config.MaxUserConnection+1)
	servermodule.AddModule(s, new(MessageHandler), 10)
	log.Infof(" %s is now running", config.AppName)
}
