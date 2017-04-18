package server

import (
	"fmt"
	"murgo/data"
)

type SessionManager struct {
	supervisor *Supervisor
	cast chan interface{}
}


func NewSessionManager (supervisor *Supervisor)(*SessionManager) {
	sessionManager := new(SessionManager)
	sessionManager.cast = make(chan interface{})
	sessionManager.supervisor = supervisor

	return sessionManager
}

func (sessionManager *SessionManager)startSessionManager(supervisor *Supervisor) {
	for{
		select {
		case castData := <-sessionManager.cast:
			sessionManager.handleCast(castData)

		}
	}
}

func (sessionManager *SessionManager)handleCast (castData interface{}) {
	//fmt.Println(" casthandler entered")

	switch t := castData.(type) {
	default:
		fmt.Printf("unexpected type %T", t)
	case *data.Message:
	//todo
	}
}