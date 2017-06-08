// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo message handler

package server

import (
	"fmt"

	"murgo/pkg/mumbleproto"

	"github.com/golang/protobuf/proto"
)

type MessageHandler struct {
	supervisor *Supervisor

	Cast chan interface{}
	Call chan interface{}
}

func NewMessageHandler(supervisor *Supervisor) *MessageHandler {
	messageHandler := new(MessageHandler)
	messageHandler.supervisor = supervisor

	messageHandler.Cast = make(chan interface{})
	return messageHandler
}

func (messageHandler *MessageHandler) startMassageHandler() {
	// panic 발생시 모든 모듈의 이 시점으로 리턴할 것
	// TODO : 일단 에러 발생 시점 파악을 위해 주석처리 이후에 슈퍼바이저에서 코드 통합 강구
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Message Handler recovered")
			messageHandler.startMassageHandler()
		}
	}()

	fmt.Println("Message Handler stared")
	for {
		select {
		case castData := <-messageHandler.Cast:
			messageHandler.handleCast(castData)

		}
	}
}

func (messageHandler *MessageHandler) handleCast(castData interface{}) {
	switch castData.(type) {
	default:
		panic("Handling cast of unexpected type in message handler")
	case *Message:
		msg := castData.(*Message)
		messageHandler.handleMassage(msg)
	}
}

func (messageHandler *MessageHandler) handleMassage(msg *Message) {

	switch msg.kind {
	case mumbleproto.MessageAuthenticate:
		messageHandler.handleAuthenticateMessage(msg)
	case mumbleproto.MessagePing:
		messageHandler.handlePingMessage(msg)
	case mumbleproto.MessageChannelRemove:
		messageHandler.handleChannelRemoveMessage(msg)
	case mumbleproto.MessageChannelState:
		messageHandler.handleChannelStateMessage(msg)
	case mumbleproto.MessageUserState:
		messageHandler.handleUserStateMessage(msg)
	case mumbleproto.MessageUserRemove:
		messageHandler.handleUserRemoveMessage(msg)
	case mumbleproto.MessageBanList:
		messageHandler.handleBanListMessage(msg)
	case mumbleproto.MessageTextMessage:
		messageHandler.handleTextMessage(msg)
	case mumbleproto.MessageACL:
		messageHandler.handleAclMessage(msg)
	case mumbleproto.MessageQueryUsers:
		messageHandler.handleQueryUsers(msg)
	case mumbleproto.MessageCryptSetup:
		messageHandler.handleCryptSetup(msg)
	case mumbleproto.MessageContextAction:
		fmt.Print("MessageContextAction from client")
	case mumbleproto.MessageUserList:
		messageHandler.handleUserList(msg)
	case mumbleproto.MessageVoiceTarget:
		messageHandler.handleVoiceTarget(msg)
	case mumbleproto.MessagePermissionQuery:
		messageHandler.handlePermissionQuery(msg)
	case mumbleproto.MessageUserStats:
		messageHandler.handleUserStatsMessage(msg)
	case mumbleproto.MessageRequestBlob:
		messageHandler.handleRequestBlob(msg)
	}
}

////Authenticate message handling
func (messageHandler *MessageHandler) handleAuthenticateMessage(msg *Message) {

	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, authenticate)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Authenticate info:", authenticate)

	// crypto setup
	client := msg.client

	client.userName = *authenticate.Username
	client.cryptState.GenerateKey()
	err = client.sendMessage(&mumbleproto.CryptSetup{
		Key:         client.cryptState.Key(),
		ClientNonce: client.cryptState.EncryptIV(),
		ServerNonce: client.cryptState.DecryptIV(),
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
	messageHandler.supervisor.cm.Cast <- &MurgoMessage{
		kind:   sendChannelList,
		client: client,
	}
	// enter the root channel as default channel
	messageHandler.supervisor.cm.Cast <- &MurgoMessage{
		kind:      userEnterChannel,
		client:    client,
		channelId: ROOT_CHANNEL,
	}

	sync := &mumbleproto.ServerSync{}
	sync.Session = proto.Uint32(uint32(client.session))
	sync.MaxBandwidth = proto.Uint32(72000)
	sync.WelcomeText = proto.String("Welcome to murgo server")
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

func (messageHandler *MessageHandler) handlePingMessage(msg *Message) {
	ping := &mumbleproto.Ping{}
	err := proto.Unmarshal(msg.buf, ping)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println("ping info:", ping, "msg id : ", msg.testCounter)
	client := msg.client
	client.sendMessage(&mumbleproto.Ping{
		Timestamp: ping.Timestamp,
		Good:      proto.Uint32(1),
		Late:      proto.Uint32(1),
		Lost:      proto.Uint32(1),
		Resync:    proto.Uint32(1),
	})
}

func (messageHandler *MessageHandler) handleUserStateMessage(msg *Message) {

	// 메시지를 보낸 유저 reset idle -> 이 부분은 통합
	userstate := &mumbleproto.UserState{}
	err := proto.Unmarshal(msg.buf, userstate)
	if err != nil {
		//
		return
	}
	fmt.Println("userstate info:", userstate, "msg id :", msg.testCounter, "from:", msg.client.userName)
	//Channel ID 필드 값이 있는 경우
	if userstate.ChannelId != nil {
		messageHandler.supervisor.cm.Cast <- &MurgoMessage{
			kind:      userEnterChannel,
			channelId: int(*userstate.ChannelId),
			client:    msg.client, // temp target
		}
	}

	clients := messageHandler.supervisor.tc
	channelManager := messageHandler.supervisor.cm

	actor, ok := clients[msg.client.Session()]
	if !ok {
		//server.Panic("Client not found in server's client map.")
		return
	}

	//actor는 메시지를 보낸 클라이언트
	//target은 메세지 패킷의 session

	target := actor
	if userstate.Session != nil {
		// target -> 메시지의 session에 해당하는 client 메시지의 대상. sender일 수도 있고 아닐 수도 있다
		target, ok = clients[*userstate.Session]
		if !ok {
			fmt.Println("Invalid session in UserState message")
			return
		}
	}

	userstate.Session = proto.Uint32(target.Session())
	userstate.Actor = proto.Uint32(actor.Session())

	tempUserState := &mumbleproto.UserState{}
	if userstate.Mute != nil {
		if actor.Session() != target.Session() {
			//can't change other users mute state
			//permission denied
			sendPermissionDenied(actor, mumbleproto.PermissionDenied_Permission)
		} else {
			// 변경
			tempUserState.Mute = userstate.Mute
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
		Name:     userstate.Name,
	}

	if userstate.ChannelId != nil {
		channelManager.Cast <- &MurgoMessage{
			kind:      broadCastChannel,
			channelId: int(*userstate.ChannelId),
			msg:       newMsg,
		}
	}

}

//구현된 핸들링 함수
func (messageHandler *MessageHandler) handleChannelStateMessage(msg *Message) {

	channelStateMsg := &mumbleproto.ChannelState{}
	err := proto.Unmarshal(msg.buf, channelStateMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ChannelState info:", channelStateMsg, "from:", msg.client.userName)
	if channelStateMsg.ChannelId == nil && channelStateMsg.Name != nil && *channelStateMsg.Temporary == true && *channelStateMsg.Parent == 0 && *channelStateMsg.Position == 0 {
		messageHandler.supervisor.cm.Cast <- &MurgoMessage{
			kind:        addChannel,
			ChannelName: *channelStateMsg.Name,
			client:      msg.client,
		}
	}
}
func (messageHandler *MessageHandler) handleUserStatsMessage(msg *Message) {
	userStats := &mumbleproto.UserStats{}
	client := msg.client
	err := proto.Unmarshal(msg.buf, userStats)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("userstats info:", userStats, "from:", msg.client.userName)
	newUserStatsMsg := &mumbleproto.UserStats{
		TcpPingAvg: proto.Float32(client.tcpPingAvg),
		TcpPingVar: proto.Float32(client.tcpPingVar),
		Opus:       proto.Bool(client.opus),
		//TODO : elapsed time 계산 과정 추가해서 idle, online 시간 추적
		//Bandwidth:
		//Onlinesecs:
		//Idlesecs:
	}
	client.sendMessage(newUserStatsMsg)
}

// protocol handling dummy for UserRemoveMessage
func (messageHandler *MessageHandler) handleUserRemoveMessage(msg *Message) {
	msgProto := &mumbleproto.UserRemove{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("userRemoveMessage info:", msgProto, "from:", msg.client.userName)
}

// protocol handling dummy for banlist message
func (messageHandler *MessageHandler) handleBanListMessage(msg *Message) {

	msgProto := &mumbleproto.BanList{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Banlist message info:", msgProto, "from:", msg.client.userName)
}

//TODO : deal with sending text message
func (messageHandler *MessageHandler) handleTextMessage(msg *Message) {
	textMsg := &mumbleproto.TextMessage{}
	err := proto.Unmarshal(msg.buf, textMsg)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("Text msg info:", textMsg, "from:", msg.client.userName)

	if textMsg.ChannelId != nil {
		newMsg := mumbleproto.TextMessage{
			Actor:     textMsg.Actor,
			Session:   textMsg.Session,
			ChannelId: textMsg.ChannelId,
			Message:   textMsg.Message,
		}
		messageHandler.supervisor.cm.Cast <- &MurgoMessage{
			kind:      broadCastChannel,
			msg:       newMsg,
			channelId: int(textMsg.ChannelId[0]),
		}
	} else if textMsg.Session != nil {

	}

}

// protocol handling dummy for Aclmessage

func (messageHandler *MessageHandler) handleAclMessage(msg *Message) {
	msgProto := &mumbleproto.ACL{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ACL info:", msgProto, "from:", msg.client.userName)

}

// protocol handling dummy for QueryUSers

func (messageHandler *MessageHandler) handleQueryUsers(msg *Message) {
	msgProto := &mumbleproto.QueryUsers{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Handle query users  :", msgProto, "from:", msg.client.userName)

}

// protocol handling dummy for CtyptSetup
func (messageHandler *MessageHandler) handleCryptSetup(msg *Message) {
	msgProto := &mumbleproto.CryptSetup{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("handle cryptsetup :", msgProto, "from:", msg.client.userName)

}

// protocol handling dummy for UserList

func (messageHandler *MessageHandler) handleUserList(msg *Message) {
	msgProto := &mumbleproto.UserList{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(" userlist :", msgProto, "from:", msg.client.userName)

}

// protocol handling dummy for Voicetarget

func (messageHandler *MessageHandler) handleVoiceTarget(msg *Message) {
	msgProto := &mumbleproto.VoiceTarget{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("voice target info:", msgProto, "from:", msg.client.userName)

}

// protocol handling dummy for PermissionQuery
func (messageHandler *MessageHandler) handlePermissionQuery(msg *Message) {
	msgProto := &mumbleproto.PermissionQuery{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("permission query info:", msgProto, "from:", msg.client.userName)

}

// protocol handling dummy for requestBlob
func (messageHandler *MessageHandler) handleRequestBlob(msg *Message) {
	msgProto := &mumbleproto.RequestBlob{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("requestblob info:", msgProto, "from:", msg.client.userName)

}

func (messageHandler *MessageHandler) handleChannelRemoveMessage(msg *Message) {
	msgProto := &mumbleproto.RequestBlob{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("channel remove info:", msgProto, "from:", msg.client.userName)
}

// TODO : permission 처리 나누어서 구현
// Send message when permission denied
func sendPermissionDenied(client *TlsClient, denyType mumbleproto.PermissionDenied_DenyType) {
	permissionDeniedMsg := &mumbleproto.PermissionDenied{
		Session: proto.Uint32(client.Session()),
		Type:    &denyType,
	}
	fmt.Println("Permission denied ", permissionDeniedMsg)
	err := client.sendMessage(permissionDeniedMsg)
	if err != nil {
		fmt.Println("Error sending messsage")
		return
	}

}
