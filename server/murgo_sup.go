// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor


package server

import (
	"fmt"
	"errors"
	"murgo/pkg/servermodule"
)


type MurgoSupervisor struct {
	servermodule.Supervisor

	sessionManager *SessionManager
	messageHandler *MessageHandler
	channelManager *ChannelManager
	tlsServer      *TlsServer
	tlsClients     map[uint32]*TlsClient
}

func (murgoSupervisor *MurgoSupervisor)Init(){

	sessionManager := NewSessionManager()
	sessionManager.GenServer = servermodule.NewGenServer(*sessionManager)

	murgoSupervisor.StartGenServer(sessionManager.GenServer)



	murgoSupervisor.tlsClients = make( map[uint32]*TlsClient)
	murgoSupervisor.messageHandler = NewMessageHandler(murgoSupervisor)
	murgoSupervisor.tlsServer = NewTlsServer(murgoSupervisor)
	murgoSupervisor.channelManager = NewChannelManager(murgoSupervisor)


}


/*
func (murgoSupervisor *MurgoSupervisor)StartSupervisor() {

	murgoSupervisor.
	murgoSupervisor.startGenServer(murgoSupervisor.sessionManager.startSessionManager)
	murgoSupervisor.startGenServer(murgoSupervisor.channelManager.StartChannelManager)
	murgoSupervisor.startGenServer(murgoSupervisor.messageHandler.startMassageHandler)
	murgoSupervisor.startGenServer(murgoSupervisor.tlsServer.startTlsServer)
}*/

func (supervisor *MurgoSupervisor)Terminate() (error){
	// todo : terminate
	return errors.New("Error")
}

func (supervisor *MurgoSupervisor) handleCast (castData interface{}) {

	switch t := castData.(type) {
	default:
		fmt.Printf(" : unexpected type %T", t)
	}
}

/*func (supervisor *Supervisor)startGenServer1(serverInterface interface{}) {
	serverInterface.(clientAction).start()
}*/

/*
func (supervisor *MurgoSupervisor)startGenServer(genServer func()) {
	fmt.Println("gen server started")

	go genServer()
}
*/

