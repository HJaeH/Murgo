package server

import (
	"fmt"

	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/pkg/sessionpool"
	APIkeys "murgo/server/util"

	"net"

	"github.com/golang/protobuf/proto"
)

type SessionManager struct {
	clientList   map[uint32]*TlsClient
	sessionPool  *sessionpool.SessionPool
	clientIdList map[uint32]*TlsClient
	servermodule.GenCallback
}

func (s *SessionManager) getClientList() map[uint32]*TlsClient {
	return s.clientList
}

//Callbacks
func (s *SessionManager) Init() {
	servermodule.RegisterAPI((*SessionManager).HandleIncomingClient, APIkeys.HandleIncomingClient)
	s.sessionPool = sessionpool.New()
	s.clientList = make(map[uint32]*TlsClient)
}

//// APIs
func (s *SessionManager) HandleIncomingClient( /*conn *net.Conn*/ a int) {
	//var conn = new(nest.Conn)
	//init tls client
	fmt.Println("jhandle cincpmig msdfjsdklasdjfk;", a)
	conn := new(net.Conn)
	/*//session := s.sessionPool.Get()*/
	session := *new(uint32)
	//client := NewTlsClient(conn, session, sessionManager.supervisor)
	client := NewTlsClient(conn, session)
	s.clientList[session] = client
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

func (s *SessionManager) BroadcastMessage(msg interface{}) {
	for _, eachClient := range s.clientList {
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
