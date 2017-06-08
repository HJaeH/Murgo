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
	clientList   map[uint32]*Client
	sessionPool  *sessionpool.SessionPool
	clientIdList map[uint32]*Client
}

func (s *SessionManager) getClientList() map[uint32]*Client {
	return s.clientList
}

//Callbacks
func (s *SessionManager) Init() {
	servermodule.RegisterAPI((*SessionManager).HandleIncomingClient, APIkeys.HandleIncomingClient)
	servermodule.RegisterAPI((*SessionManager).BroadcastMessage, APIkeys.BroadcastMessage)
	servermodule.RegisterAPI((*SessionManager).SetUserOption, APIkeys.SetUserOption)

	s.sessionPool = sessionpool.New()
	s.clientList = make(map[uint32]*Client)
}

//// APIs
func (s *SessionManager) HandleIncomingClient(conn net.Conn) {
	//conn = (*net.Conn)conn
	//var conn = new(nest.Conn)
	//init tls client

	session := s.sessionPool.Get()
	//client := NewTlsClient(conn, session, sessionManager.supervisor)
	/*a :=
	a = conn.*/
	client := NewTlsClient(&conn, session)
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
	servermodule.Cast(APIkeys.Receive, client)
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

func (s *SessionManager) SetUserOption(userState *mumbleproto.UserState) {

	actor, ok := s.clientList[*userState.Actor]
	if !ok {
		//server.Panic("Client not found in server's client map.")
		return
	}

	//actor는 메시지를 보낸 클라이언트
	//target은 메세지 패킷의 session 값; 메시지의 대상

	target := actor
	if userState.Session != nil {
		// target -> 메시지의 session에 해당하는 client 메시지의 대상. sender일 수도 있고 아닐 수도 있다
		target, ok = s.clientList[*userState.Session]
		if !ok {
			fmt.Println("Invalid session in UserState message")
			return
		}
	}

	userState.Session = proto.Uint32(target.Session())
	userState.Actor = proto.Uint32(actor.Session())

	tempUserState := &mumbleproto.UserState{}
	if userState.Mute != nil {
		if actor.Session() != target.Session() {
			//can't change other users mute state
			//permission denied
			sendPermissionDenied(actor, mumbleproto.PermissionDenied_Permission)
		} else {
			// 변경
			tempUserState.Mute = userState.Mute
		}
	} else {
		if actor.Session() != target.Session() {
			if actor.mute == false {

			}
		}
	}

	newMsg := &mumbleproto.UserState{
		Deaf:     proto.Bool(false),
		SelfDeaf: proto.Bool(false),
		Name:     userState.Name,
	}

	if userState.ChannelId != nil {
		servermodule.Cast(APIkeys.BroadcastChannel, int(*userState.ChannelId), newMsg)
	}
}

// todo : cast time.Time type to number type or overloading
/*
func elapsed(prev time.Time, now time.Time)(time.Time){
	return now - prev
}*/
