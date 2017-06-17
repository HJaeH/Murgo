package server

import (
	"fmt"

	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/pkg/sessionpool"
	"murgo/server/util/apikeys"
	"murgo/server/util/log"

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
	servermodule.RegisterAPI((*SessionManager).HandleIncomingClient, apikeys.HandleIncomingClient)
	servermodule.RegisterAPI((*SessionManager).BroadcastMessage, apikeys.BroadcastMessage)
	//servermodule.RegisterAPI((*SessionManager).SetUserOption, apikeys.SetUserOption)
	servermodule.RegisterAPI((*SessionManager).RemoveClient, apikeys.RemoveClient)
	servermodule.RegisterAPI((*SessionManager).SendMessages, apikeys.SendMessages)

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
		Release:     proto.String(config.AppName),
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
		err = s.handleAuthenticate(msg)
	}

	servermodule.AsyncCall(apikeys.Receive, client)
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
			actor.sendPermissionDenied(mumbleproto.PermissionDenied_Permission)
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
		servermodule.AsyncCall(apikeys.BroadcastChannel, *userState.ChannelId, newUserState)
	}
}

func (s *SessionManager) SendMessages(sessions []uint32, msg interface{}) {
	for _, session := range sessions {
		fmt.Println(session)
		if eachClient, ok := s.clientList[session]; ok {
			eachClient.sendMessage(msg)
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

func (s *SessionManager) handleAuthenticate(msg *Message) error {

	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, authenticate)
	if err != nil {
		return err
	}
	client := msg.client
	newUserName := *authenticate.Username
	if !s.isValidName(newUserName) {
		client.Disconnect()
		return log.Error("Duplicated username")
	}
	client.UserName = newUserName

	client.crypt.GenerateKey()
	cryptMsg := &mumbleproto.CryptSetup{
		Key:         client.crypt.Key(),
		ClientNonce: client.crypt.EncryptIV(),
		ServerNonce: client.crypt.DecryptIV(),
	}
	if err = client.sendMessage(cryptMsg); err != nil {
		return err
	}

	client.codecs = authenticate.CeltVersions
	if len(client.codecs) == 0 {
		//todo : no codec msg case
	}

	//send codec version
	codecMsg := &mumbleproto.CodecVersion{
		Alpha:       proto.Int32(-2147483637),
		Beta:        proto.Int32(-2147483632),
		PreferAlpha: proto.Bool(false),
		Opus:        proto.Bool(true),
	}
	if err = client.sendMessage(codecMsg); err != nil {
		return err
	}

	/// send channel state
	servermodule.AsyncCall(apikeys.SendChannelList, client)
	// enter the root channel as default channel
	servermodule.AsyncCall(apikeys.EnterChannel, ROOT_CHANNEL, client)

	sync := &mumbleproto.ServerSync{
		Session:      proto.Uint32(uint32(client.session)),
		MaxBandwidth: proto.Uint32(72000),
		WelcomeText:  proto.String(config.WelComeMessage),
		Permissions:  proto.Uint64(uint64(acl.AllPermissions)),
	}
	if err := client.sendMessage(sync); err != nil {
		return err
	}

	serverConfigMsg := &mumbleproto.ServerConfig{
		AllowHtml:     proto.Bool(true),
		MessageLength: proto.Uint32(128),
		MaxBandwidth:  proto.Uint32(240000),
	}
	if err := client.sendMessage(serverConfigMsg); err != nil {
		return err
	}

	return nil
}

// todo : cast time.Time type to number type or overloading
/*
func elapsed(prev time.Time, now time.Time)(time.Time){
	return now - prev
}*/
