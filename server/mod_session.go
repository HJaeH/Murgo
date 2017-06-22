package server

import (
	"fmt"

	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/pkg/sessionpool"
	"murgo/server/util/log"

	"net"

	"io"

	"murgo/server/util/event"

	"murgo/server/util/crypt"

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
func (s *SessionManager) client(session uint32) *Client {
	client, ok := s.clientList[session]
	if ok != true {
		log.ErrorP("invalid session")
	}
	return client
}

func (s *SessionManager) RemoveClient(client *Client) {
	//delete from session list
	delete(s.clientList, client.session)
	//delete from channel
	channel := client.Channel
	if channel != nil {
		channel.leave(client)
		if channel.IsEmpty() {
			servermodule.Call(event.RemoveChannel, channel)
		}

	}

	s.sessionPool.Reclaim(client.Session())
}

//Callbacks
func (s *SessionManager) Init() {
	servermodule.RegisterAPI((*SessionManager).HandleIncomingClient, event.HandleIncomingClient)
	servermodule.RegisterAPI((*SessionManager).BroadcastMessage, event.BroadcastMessage)
	servermodule.RegisterAPI((*SessionManager).RemoveClient, event.RemoveClient)
	servermodule.RegisterAPI((*SessionManager).SendMultipleMessage, event.SendMultipleMessages)
	servermodule.RegisterAPI((*SessionManager).GiveSpeakAbility, event.GiveSpeakAbility)

	s.sessionPool = sessionpool.New()
	s.clientList = make(map[uint32]*Client)
}

//// APIs
func (s *SessionManager) HandleIncomingClient(conn net.Conn) {
	session := s.sessionPool.Get()
	client := newClient(&conn, session)
	s.clientList[session] = client
	// send version information
	version := &mumbleproto.Version{
		Version:     proto.Uint32(0x10205),
		Release:     proto.String(config.AppName),
		CryptoModes: crypt.SupportedModes(),
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
	servermodule.AsyncCall(event.Receive, client)
}

func (s *SessionManager) BroadcastMessage(msg interface{}) {
	for _, eachClient := range s.clientList {
		eachClient.sendMessage(msg)
	}
}

func (s *SessionManager) SendMultipleMessage(sessions []uint32, msg interface{}) error {

	for _, session := range sessions {
		if eachClient, ok := s.clientList[session]; ok {
			eachClient.sendMessage(msg)
		} else {
			panic("session not exist")
		}
	}

	return nil
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
	servermodule.Call(event.SendChannelList, client)
	// enter the root channel as default channel
	servermodule.Call(event.EnterChannel, ROOT_CHANNEL, client)

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

func (s *SessionManager) GiveSpeakAbility(userState *mumbleproto.UserState) error {

	actor := s.client(userState.GetActor())
	target := s.client(userState.GetSession())
	//can't change other person's right to speak
	if userState.GetMute() == true {
		if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
			return err
		}
	}

	// give actor's right to talk to the target
	if userState.GetMute() == false {
		//only when actor has the right
		if target.existUsableMic &&
			target.existUsableSpeaker &&
			actor.mute == false {
			actor.mute = true
			newUserState := &mumbleproto.UserState{
				Session: proto.Uint32(actor.Session()),
				Actor:   proto.Uint32(actor.Session()),
				Mute:    proto.Bool(true),
			}

			servermodule.AsyncCall(event.BroadcastChannel, actor.Channel.Id, newUserState)

			// give it to target
			target.mute = true
			newUserState.Session = proto.Uint32(target.Session())
			newUserState.Actor = proto.Uint32(target.Session())
			newUserState.Mute = proto.Bool(false)
			servermodule.AsyncCall(event.BroadcastChannel, target.Channel.Id, newUserState)
		}
		return nil
	}
	return nil
}
