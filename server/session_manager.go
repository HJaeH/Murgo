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
	supervisor *Supervisor

	clientList map[uint32]*TlsClient

	sessionPool *sessionpool.SessionPool

	Cast chan interface{}
	Call chan interface{}
}

func NewSessionManager(supervisor *Supervisor) *SessionManager {
	sessionManager := new(SessionManager)
	sessionManager.Cast = make(chan interface{})
	sessionManager.supervisor = supervisor
	sessionManager.sessionPool = sessionpool.New()
	sessionManager.clientList = make(map[uint32]*TlsClient)
	return sessionManager
}

const (
	broadcastMessage uint16 = iota
	handleIncomingClient
)

func (sessionManager *SessionManager) startSessionManager() {

	// TODO : panic 발생시 모든 모듈의 이 시점으로 리턴할 것
	// TODO : 일단 에러 발생 시점 파악을 위해 주석처리 이후에 슈퍼바이저에서 코드 통합 강구
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Session manager recovered")
			sessionManager.startSessionManager()
		}
	}()

	for {
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

func (sessionManager *SessionManager) handleCast(castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch murgoMsg.kind {
	default:
		fmt.Printf("unexpected type %T", murgoMsg.kind)
	case broadcastMessage:
		sessionManager.broadcastMessage(murgoMsg.msg)
	case handleIncomingClient:
		sessionManager.handleIncomingClient(murgoMsg.conn)
	}
}

func (sessionManaser *SessionManager) handleIncomingClient(conn *net.Conn) {

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

func (sessionManager *SessionManager) broadcastMessage(msg interface{}) {
	for _, eachClient := range sessionManager.clientList {
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

// todo : cast time.Time type to number type or check overlaoding
/*
func elapsed(prev time.Time, now time.Time)(time.Time){
	return now - prev
}*/
