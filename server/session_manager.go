package server

import (
	"fmt"
	"net"

	"mumble.info/grumble/pkg/mumbleproto"
	"murgo/config"

	"github.com/golang/protobuf/proto"
	"mumble.info/grumble/pkg/sessionpool"
)

type SessionManager struct {
	supervisor *Supervisor

	clientList map[uint32] *TlsClient

	sessionPool *sessionpool.SessionPool



	Cast chan interface{}
	Call chan interface{}

}




func NewSessionManager (supervisor *Supervisor)(*SessionManager) {
	sessionManager := new(SessionManager)
	sessionManager.Cast = make(chan interface{})
	sessionManager.supervisor = supervisor
	sessionManager.sessionPool = sessionpool.New()
	sessionManager.clientList = make(map[uint32] *TlsClient)
	return sessionManager
}


const (
	broadcastMessage uint16 = iota
	handleIncomingClient


)


func (sessionManager *SessionManager)startSessionManager() {

	// TODO : panic 발생시 모든 모듈의 이 시점으로 리턴할 것
	// TODO : 일단 에러 발생 시점 파악을 위해 주석처리 이후에 슈퍼바이저에서 코드 통합 강구
	/*defer func(){
		if err:= recover(); err!= nil{
			fmt.Println("Session manager recovered")
			sessionManager.startSessionManager()
		}
	}()
*/

	for{
		select {
		case castData := <-sessionManager.Cast:
			sessionManager.handleCast(castData)

		}
	}
}
//todo 분기 함수포인터로 바로 접근 하는 방법.
/*
func (sessionManager *SessionManager)ha(msg *Message) {
	if
	handle(sessionManager.clientList)
}

func (sessionManager *SessionManager)handle(F func(int, int)() ) {

	F(3, 4)
}
*/



func (sessionManager *SessionManager)handleCast(castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch  murgoMsg.kind {
	default:
		fmt.Printf("unexpected type %T", murgoMsg.kind)
	case broadcastMessage:
		sessionManager.broadcastMessage(murgoMsg.msg)
	case handleIncomingClient:
		sessionManager.handleIncomingClient(murgoMsg.conn)
	}
}


func (sessionManaser *SessionManager)handleIncomingClient(conn *net.Conn){

	//init tls client
	session := sessionManaser.sessionPool.Get()
	client := NewTlsClient(conn, session)

	sessionManaser.clientList[session] = client
	// send version information
	version := &mumbleproto.Version{
		Version:     proto.Uint32(0x10205),
		Release:     proto.String("Murgo"),
		CryptoModes: config.SupportedModes(),
	}
	err := client.sendMessage(version)
	if err != nil {
		fmt.Println("Error sending message to client")
	}

	//create client message receive loop as gen server
	// TODO : the start time need to be pushed back - after check duplication
	// TODO : but the work is conducted in authenticate which is running in message accepting loop
	sessionManaser.supervisor.startGenServer(client.recvLoop)
}


func (sessionManager *SessionManager) broadcastMessage(msg interface{}){
	fmt.Println("broad cast")
	for _, eachClient := range sessionManager.clientList {
		/*if client.state < StateClientAuthenticated {
			continue
		}*/
		eachClient.sendMessage(msg)
	}
}
// todo : cast time.Time type to number type or check overlaoding
/*
func elapsed(prev time.Time, now time.Time)(time.Time){
	return now - prev
}*/
