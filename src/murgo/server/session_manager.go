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

func (sessionManager *SessionManager)startSessionManager(supervisor *Supervisor) {
	for{
		select {
		case castData := <-sessionManager.Cast:
			sessionManager.handleCast(castData)

		}
	}
}

func (sessionManager *SessionManager)handleCast (castData interface{}) {
	//fmt.Println(" casthandler entered")

	switch t := castData.(type) {
	default:
		fmt.Printf("unexpected type %T", t)
	case *Message:
	//todo
	}
}