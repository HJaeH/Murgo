package server

import (
	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/pkg/servermodule/log"
	"murgo/pkg/sessionpool"

	"net"

	"io"

	"murgo/server/util/event"

	"murgo/server/util/crypt"

	"errors"

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
		log.Error("invalid session")
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

func (s *SessionManager) HandleIncomingClient(conn net.Conn) {

	session := s.sessionPool.Get()
	client := newClient(&conn, session)
	s.clientList[session] = client

	//check current client count
	if s.totalClientCount() >= config.MaxUserConnection {
		if err := client.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
			log.Error(err)

		}
		client.Disconnect()
		log.Error("Server is full")
		return
	}

	// Version exchange
	versionMsg := &mumbleproto.Version{
		Version:     proto.Uint32(0x10205),
		Release:     proto.String(config.AppName),
		CryptoModes: crypt.SupportedModes(),
	}
	if err := client.sendMessage(versionMsg); err != nil {
		client.Disconnect()
		log.Error(err)
		return
	}

	if msg, err := client.readProtoMessage(); err != nil {
		client.Disconnect()
		if err != io.EOF {
			log.Error(err)
		}
		return
	} else {
		if msg.kind == mumbleproto.MessageVersion {
			version := &mumbleproto.Version{}
			err := proto.Unmarshal(msg.buf, version)
			if err != nil {
				log.Error(err)
				return
			}
			client.version = *version.Version

		} else {
			log.Error("Unexpected connection setup protocol")
			client.Disconnect()
			return
		}
	}

	// generate random key
	client.crypt.GenerateKey()
	cryptMsg := &mumbleproto.CryptSetup{
		Key:         client.crypt.Key(),
		ClientNonce: client.crypt.EncryptIV(),
		ServerNonce: client.crypt.DecryptIV(),
	}
	if err := client.sendMessage(cryptMsg); err != nil {
		client.Disconnect()
		log.Error(err)
		return
	}

	// Authentication
	if msg, err := client.readProtoMessage(); err != nil {
		client.Disconnect()
		if err != io.EOF {
			log.Error(err)
		}
		return
	} else {
		if msg.kind == mumbleproto.MessageAuthenticate {
			if err := s.handleAuthenticate(msg); err != nil {
				client.Disconnect()
				log.Error(err)
				return
			}

		} else {
			log.Error("Invalid connection setup protocol")
			client.Disconnect()
			return
		}
	}

	//run receive loop
	servermodule.AsyncCall(event.Receive, client)
}

func (s *SessionManager) BroadcastMessage(msg interface{}) {
	for _, eachClient := range s.clientList {
		if err := eachClient.sendMessage(msg); err != nil {
			log.Error("Error sending message")
		}
	}
}

func (s *SessionManager) SendMultipleMessage(sessions []uint32, msg interface{}) error {
	for _, session := range sessions {
		if eachClient, ok := s.clientList[session]; ok {
			if err := eachClient.sendMessage(msg); err != nil {
				log.Error("Error sending message")
			}
		} else {
			log.Error("Session id doesn't exist")
			return errors.New("Session id doesn't exist")
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

	client := msg.client

	//Dealing with authenticate message
	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, authenticate)
	if err != nil {
		return err
	}
	//check username validation
	userName := authenticate.GetUsername()
	if !s.isValidName(userName) {
		if err := client.sendPermissionDenied(mumbleproto.PermissionDenied_UserName); err != nil {
			errors.New("Error sending message")
		}
		client.Disconnect()
		return errors.New("Duplicated username")
	}
	client.UserName = userName
	//set codecs
	client.codecs = authenticate.GetCeltVersions()

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

	//send server sync
	sync := &mumbleproto.ServerSync{
		Session:      proto.Uint32(uint32(client.session)),
		MaxBandwidth: proto.Uint32(config.MaxBandwidth),
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
			servermodule.Call(event.BroadcastChannel, actor.Channel.Id, newUserState)

			// give it to target
			target.mute = false
			newUserState = &mumbleproto.UserState{
				Session: proto.Uint32(target.Session()),
				Actor:   proto.Uint32(target.Session()),
				Mute:    proto.Bool(false),
			}
			//let other clients know the user's state
			servermodule.Call(event.BroadcastChannel, target.Channel.Id, newUserState)
			return nil
		} else {
			target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission)
			return nil
		}
	}
	return errors.New("Invalid message data")
}

func (s *SessionManager) totalClientCount() int {

	return len(s.clientList)
}

// callback
func (s *SessionManager) Init() {
	s.sessionPool = sessionpool.New()
	s.clientList = make(map[uint32]*Client)

	//add event
	servermodule.EventRegister((*SessionManager).HandleIncomingClient, event.HandleIncomingClient)
	servermodule.EventRegister((*SessionManager).BroadcastMessage, event.BroadcastMessage)
	servermodule.EventRegister((*SessionManager).RemoveClient, event.RemoveClient)
	servermodule.EventRegister((*SessionManager).SendMultipleMessage, event.SendMultipleMessages)
	servermodule.EventRegister((*SessionManager).GiveSpeakAbility, event.GiveSpeakAbility)

}
