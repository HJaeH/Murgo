package server

import (
	"fmt"

	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/pkg/sessionpool"
	APIkeys "murgo/server/util"

	"net"

	"io"

	"github.com/golang/protobuf/proto"
	"mumble.info/grumble/pkg/acl"
)

type SessionManager struct {
	clientList  map[uint32]*Client
	sessionPool *sessionpool.SessionPool
}

func (s *SessionManager) getClientList() map[uint32]*Client {
	return s.clientList
}
func (s *SessionManager) getClient(session uint32) *Client {
	client, ok := s.clientList[session]
	if ok != true {
		panic("invalid session")
	}
	return client
}

func (s *SessionManager) RemoveClient(client *Client) {
	//delete from session list
	delete(s.clientList, client.session)
	//delete from channel
	channel := client.Channel
	if channel != nil {
		channel.removeClient(client)
	}
	s.sessionPool.Reclaim(client.Session())
}

//Callbacks
func (s *SessionManager) Init() {
	servermodule.RegisterAPI((*SessionManager).HandleIncomingClient, APIkeys.HandleIncomingClient)
	servermodule.RegisterAPI((*SessionManager).BroadcastMessage, APIkeys.BroadcastMessage)
	servermodule.RegisterAPI((*SessionManager).SetUserOption, APIkeys.SetUserOption)
	servermodule.RegisterAPI((*SessionManager).RemoveClient, APIkeys.RemoveClient)
	servermodule.RegisterAPI((*SessionManager).SendMessages, APIkeys.SendMessages)

	s.sessionPool = sessionpool.New()
	s.clientList = make(map[uint32]*Client)
}

//// APIs
func (s *SessionManager) HandleIncomingClient(conn net.Conn) {

	session := s.sessionPool.Get()
	client := newClient(&conn, session)
	fmt.Println(session, "is connected")
	s.clientList[session] = client
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

	msg, err := client.readProtoMessage()
	if err != nil {
		if err != nil {
			if err == io.EOF {
				client.Disconnect()
			} else {
				panic(err)
			}
			return
		}
	}
	if msg.kind == mumbleproto.MessageVersion {
		version := &mumbleproto.Version{}
		err := proto.Unmarshal(msg.buf, version)
		if err != nil {
			fmt.Println(err)
			return
		}

		client.version = *version.Version
		fmt.Println(*version.Version)
	}

	msg, err = client.readProtoMessage()
	if err != nil {
		if err != nil {
			if err == io.EOF {
				client.Disconnect()
			} else {
				panic(err)
			}
			return
		}
	}

	if msg.kind == mumbleproto.MessageAuthenticate {
		s.handleAuthenticate(msg)
	}

	servermodule.AsyncCall(APIkeys.Receive, client)
}

func (s *SessionManager) BroadcastMessage(msg interface{}) {
	for _, eachClient := range s.clientList {
		/*if client.state < StateClientAuthenticated {
			continue
		}*/
		eachClient.sendMessage(msg)
	}
}

func (s *SessionManager) SetUserOption(client *Client, userState *mumbleproto.UserState) {

	//actor는 메시지를 보낸 클라이언트
	actor, ok := s.clientList[client.Session()]
	if !ok {
		panic("Client not found in server's client map.")
		return
	}

	//target이 없으면 actor가 target
	target := actor
	if userState.Session != nil {
		target, ok = s.clientList[*userState.Session]
		if !ok {
			fmt.Println("Invalid session in UserState message")
			return
		}
	}

	userState.Session = proto.Uint32(target.Session())
	userState.Actor = proto.Uint32(actor.Session())

	newUserState := &mumbleproto.UserState{
		Deaf:     proto.Bool(false),
		SelfDeaf: proto.Bool(false),
		Name:     userState.Name,
	}
	if userState.Mute != nil {
		if actor.Session() != target.Session() {
			//can't change other users mute state
			//permission denied
			sendPermissionDenied(actor, mumbleproto.PermissionDenied_Permission)
		} else {
			// 변경
			newUserState.Mute = userState.Mute
		}
	} else {
		if actor.Session() != target.Session() {
			if actor.mute == false {

			}
		}
	}

	if userState.ChannelId != nil {
		servermodule.AsyncCall(APIkeys.BroadcastChannel, *userState.ChannelId, newUserState)
	}
}

func (s *SessionManager) SendMessages(sessions []uint32, msg interface{}) {
	for _, session := range sessions {
		fmt.Println(session)
		if client, ok := s.clientList[session]; ok {
			client.sendMessage(msg)
		} else {
			panic("session not exist")
		}
	}
}
func (s *SessionManager) isValidName(userName string) bool {

	for _, eachClient := range s.clientList {
		if eachClient.UserName == userName {
			return false
		}
	}
	return true
}

func (s *SessionManager) handleAuthenticate(msg *Message) {

	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, authenticate)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := msg.client
	newUserName := *authenticate.Username
	if !s.isValidName(newUserName) {
		client.Disconnect()
		return
	}
	client.UserName = newUserName

	client.crypt.GenerateKey()
	err = client.sendMessage(&mumbleproto.CryptSetup{
		Key:         client.crypt.Key(),
		ClientNonce: client.crypt.EncryptIV(),
		ServerNonce: client.crypt.DecryptIV(),
	})
	if err != nil {
		fmt.Println("error sending msg")
	}
	client.codecs = authenticate.CeltVersions
	if len(client.codecs) == 0 {
		//todo : no codec msg case
	}

	//send codec version
	err = client.sendMessage(&mumbleproto.CodecVersion{
		Alpha:       proto.Int32(-2147483637),
		Beta:        proto.Int32(-2147483632),
		PreferAlpha: proto.Bool(false),
		Opus:        proto.Bool(true),
	})

	if err != nil {
		fmt.Println("error sending codec version")
		return
	}
	/// send channel state
	servermodule.AsyncCall(APIkeys.SendChannelList, client)
	// enter the root channel as default channel
	servermodule.AsyncCall(APIkeys.EnterChannel, ROOT_CHANNEL, client)

	sync := &mumbleproto.ServerSync{}
	sync.Session = proto.Uint32(uint32(client.session))
	sync.MaxBandwidth = proto.Uint32(72000)
	sync.WelcomeText = proto.String("Welcome to murgo server")
	sync.Permissions = proto.Uint64(uint64(acl.AllPermissions))
	if err := client.sendMessage(sync); err != nil {
		fmt.Println("error sending message")
		return
	}

	serverConfigMsg := &mumbleproto.ServerConfig{
		AllowHtml:     proto.Bool(true),
		MessageLength: proto.Uint32(128),
		MaxBandwidth:  proto.Uint32(240000),
	}
	if err := client.sendMessage(serverConfigMsg); err != nil {
		fmt.Println("error sending message")
		return
	}
}

// todo : cast time.Time type to number type or overloading
/*
func elapsed(prev time.Time, now time.Time)(time.Time){
	return now - prev
}*/
