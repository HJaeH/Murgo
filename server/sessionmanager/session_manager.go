package sessionmanager

import (
	"fmt"
	"net"

	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/sessionpool"
	"murgo/pkg/servermodule"

	"github.com/golang/protobuf/proto"
	"murgo/server"
)

const SM = "sessionManager"
const (
	broadcastMessage uint16 = iota
	handleIncomingClient


)
type SessionManager struct {

	clientList map[uint32] *server.TlsClient
	sessionPool *sessionpool.SessionPool

	//Cast chan interface{}
	//Call chan interface{}

}

func StartLink()(*SessionManager){
	sessionManager := new(SessionManager)

	servermodule.StartLinkGenserver(sessionManager)
	return sessionManager
	//sessionmanager는 genserver를 실행하고 callback으로 sessionManager.init을 실행함
}
func newSessionManager() *SessionManager{
	sessionManager := new(SessionManager)
	sessionManager.clientList = make(map[uint32] *server.TlsClient)
	sessionManager.sessionPool = sessionpool.New()
	return sessionManager
}

/*func NewSessionManager(parent interface{}) *SessionManager{
	sessionManager := new(SessionManager)

	sessionManager.StartLink(parent)
	return

}*/


func (sessionManager *SessionManager) handleCast(){
	servermodule.Cast(SM, &servermodule.CastMessage{})

}



//todo : genserver pkg적용 되면 지울 것,
/*
func (sessionManager *SessionManager)startSessionManager() {
	defer func(){
		if err:= recover(); err!= nil {
			fmt.Println("Session manager recovered")
			sessionManager.startSessionManager()
		}
	}()

	for{
		select {
		case castData := <-sessionManager.Cast:
			sessionManager.handleCast(castData)

		}
	}
}
*/

/*

func (sessionManager *SessionManager)handleCast(castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch  murgoMsg.Kind {
	default:
		fmt.Printf("unexpected type %T", murgoMsg.Kind)
	case broadcastMessage:
		sessionManager.broadcastMessage(murgoMsg.Msg)
	case handleIncomingClient:
		sessionManager.handleIncomingClient(murgoMsg.Conn)
	}
}

*/

func (sessionManager *SessionManager)handleIncomingClient(conn *net.Conn){

	//init tls client
	session := sessionManager.sessionPool.Get()
	//client := NewTlsClient(conn, session, sessionManager.supervisor)
	client := server.NewTlsClient(conn, session)
	sessionManager.clientList[session] = client
	// send version information
	version := &mumbleproto.Version{
		Version:     proto.Uint32(0x10205),
		Release:     proto.String("Murgo"),
		CryptoModes: config.SupportedModes(),
	}
	err := client.SendMessage(version)
	if err != nil {
		fmt.Println("Error sending message to client")
	}

	//create client message receive loop as gen server
	// TODO : the start time need to be pushed back - after check duplication
	// TODO : but the work is conducted in authenticate which is running in message accepting loop
	//sessionManager.supervisor.startGenServer(client.recvLoop)
	//servermodule.Supervisor.StartGenServer()


}


func (sessionManager *SessionManager) broadcastMessage(msg interface{}){
	for _, eachClient := range sessionManager.clientList {
		/*if client.state < StateClientAuthenticated {
			continue
		}*/
		eachClient.SendMessage(msg)
	}
}
// todo : cast time.Time type to number type or overloading
/*
func elapsed(prev time.Time, now time.Time)(time.Time){
	return now - prev
}*/


//callbacks
func (sessionManager *SessionManager)init(){
	sessionManager.sessionPool = sessionpool.New()
	sessionManager.clientList = make(map[uint32] *server.TlsClient)
}
