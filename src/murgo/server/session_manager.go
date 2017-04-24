package server

import (
	"fmt"
)

type SessionManager struct {
	supervisor *Supervisor
	clientList map[uint32] *TlsClient



	Cast chan interface{}
	Call chan interface{}

}


func NewSessionManager (supervisor *Supervisor)(*SessionManager) {
	sessionManager := new(SessionManager)
	sessionManager.Cast = make(chan interface{})
	sessionManager.supervisor = supervisor

	return sessionManager
}


const (
	broadcastMessage uint16 = iota

)


func (sessionManager *SessionManager)startSessionManager() {
	for{
		select {
		case castData := <-sessionManager.Cast:
			sessionManager.handleCast(castData)

		}
	}
}

func (sessionManager *SessionManager)handleCast(castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch  murgoMsg.kind {
	default:
		fmt.Printf("unexpected type sm")
	case broadcastMessage:
		sessionManager.broadcastMessage(murgoMsg.msg)
	}
}

func (sessionManager *SessionManager) broadcastMessage(msg interface{}){
	for _, eachClient := range sessionManager.clientList {
		/*if client.state < StateClientAuthenticated {
			continue
		}*/
		eachClient.sendMessage(msg)
	}
}

/*
func elapsed(prev time.Time, now time.Time)(time.Time){
	return now - prev
}*/
