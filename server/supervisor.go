// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor


package server

import (
	"fmt"
	"errors"
)


type Supervisor struct {
	sm *SessionManager
	mh *MessageHandler
	cm *ChannelManager
	ts *TlsServer
	tc map[uint32]*TlsClient
}


func NewSupervisor() (*Supervisor){


	supervisor := new(Supervisor)

	supervisor.tc = make( map[uint32]*TlsClient)


	supervisor.mh = NewMessageHandler(supervisor)
	supervisor.ts = NewTlsServer(supervisor)
	supervisor.cm = NewChannelManager(supervisor)
	supervisor.sm = NewSessionManager(supervisor)

	return supervisor
}

func (supervisor *Supervisor)StartSupervisor() {


	supervisor.startGenServer(supervisor.sm.startSessionManager)
	supervisor.startGenServer(supervisor.cm.startChannelManager)
	supervisor.startGenServer(supervisor.mh.startMassageHandler)
	supervisor.startGenServer(supervisor.ts.startTlsServer)
}

func (supervisor *Supervisor)Terminate() ( error){
	// todo : terminate
	return errors.New("Error")
}

func (supervisor *Supervisor) handleCast (castData interface{}) {

	switch t := castData.(type) {
	default:
		fmt.Printf(" : unexpected type %T", t)
	}
}

/*func (supervisor *Supervisor)startGenServer1(serverInterface interface{}) {
	serverInterface.(clientAction).start()
}*/

func (supervisor *Supervisor)startGenServer(genServer func()) {
	fmt.Println("gen server started")

	go genServer()
}

