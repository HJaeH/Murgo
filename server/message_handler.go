package server

import (
	"fmt"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	APIkeys "murgo/server/util"

	"github.com/golang/protobuf/proto"
)

type MessageHandler struct{}

func (messageHandler *MessageHandler) HandleMassage(msg *Message) {
	switch msg.kind {
	case mumbleproto.MessageAuthenticate:
		messageHandler.handleAuthenticateMessage(msg)
	case mumbleproto.MessagePing:
		messageHandler.handlePingMessage(msg)
	case mumbleproto.MessageChannelRemove:
		messageHandler.handleChannelRemoveMessage(msg)
	case mumbleproto.MessageChannelState:
		messageHandler.handleChannelStateMessage(msg)
	/*case mumbleproto.MessageUserState:
	messageHandler.handleUserStateMessage(msg)*/
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
func (m *MessageHandler) handleAuthenticateMessage(msg *Message) {

	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, authenticate)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("auth message:", authenticate, "msg id :", msg.testCounter, "from:", msg.client.UserName)

	// crypto setup
	client := msg.client

	client.UserName = *authenticate.Username
	client.cryptState.GenerateKey()
	err = client.SendMessage(&mumbleproto.CryptSetup{
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
	err = client.SendMessage(&mumbleproto.CodecVersion{
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
	servermodule.Cast(APIkeys.SendChannelList, client)
	// enter the root channel as default channel
	fmt.Println("==========")
	servermodule.Cast(APIkeys.EnterChannel, ROOT_CHANNEL, client)

	sync := &mumbleproto.ServerSync{}
	sync.Session = proto.Uint32(uint32(client.session))
	sync.MaxBandwidth = proto.Uint32(72000)
	sync.WelcomeText = proto.String("Welcome to murgo server")
	if err := client.SendMessage(sync); err != nil {
		fmt.Println("error sending message")
		return
	}

	serverConfigMsg := &mumbleproto.ServerConfig{
		AllowHtml:     proto.Bool(true),
		MessageLength: proto.Uint32(128),
		MaxBandwidth:  proto.Uint32(240000),
	}
	if err := client.SendMessage(serverConfigMsg); err != nil {
		fmt.Println("error sending message")
		return
	}
}

func (m *MessageHandler) handlePingMessage(msg *Message) {
	ping := &mumbleproto.Ping{}
	err := proto.Unmarshal(msg.buf, ping)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ping message:", ping, "msg id :", msg.testCounter, "from:", msg.client.UserName)

	//fmt.Println("ping info:", ping, "msg id : ", msg.testCounter)
	client := msg.client
	client.SendMessage(&mumbleproto.Ping{
		Timestamp: ping.Timestamp,
		Good:      proto.Uint32(1),
		Late:      proto.Uint32(1),
		Lost:      proto.Uint32(1),
		Resync:    proto.Uint32(1),
	})
}

func (messageHandler *MessageHandler) handleUserStateMessage(msg *Message) {

	// 메시지를 보낸 유저 reset idle -> 이 부분은 통합
	userState := &mumbleproto.UserState{}
	err := proto.Unmarshal(msg.buf, userState)
	if err != nil {
		//
		return
	}
	fmt.Println("userstate info:", userState, "msg id :", msg.testCounter, "from:", msg.client.UserName)
	//Channel ID 필드 값이 있는 경우
	if userState.ChannelId != nil {
		servermodule.Cast(APIkeys.EnterChannel, int(*userState.ChannelId), msg.client)
	}

	servermodule.Cast(APIkeys.SetUserOption, userState)

}

//구현된 핸들링 함수
func (m *MessageHandler) handleChannelStateMessage(tempMsg interface{}) {
	msg := tempMsg.(*Message)
	channelStateMsg := &mumbleproto.ChannelState{}
	err := proto.Unmarshal(msg.buf, channelStateMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ChannelState info:", channelStateMsg, "from:", msg.client.UserName)
	if channelStateMsg.ChannelId == nil && channelStateMsg.Name != nil && *channelStateMsg.Temporary == true && *channelStateMsg.Parent == 0 && *channelStateMsg.Position == 0 {
		servermodule.Call(APIkeys.AddChannel, *channelStateMsg.Name, msg.client)
	}
}
func (m *MessageHandler) handleUserStatsMessage(msg *Message) {
	userStats := &mumbleproto.UserStats{}
	client := msg.client
	err := proto.Unmarshal(msg.buf, userStats)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("userstats info:", userStats, "from:", msg.client.UserName)
	newUserStatsMsg := &mumbleproto.UserStats{
		TcpPingAvg: proto.Float32(client.tcpPingAvg),
		TcpPingVar: proto.Float32(client.tcpPingVar),
		Opus:       proto.Bool(client.opus),
		//TODO : elapsed time 계산 과정 추가해서 idle, online 시간 추적
		//Bandwidth:
		//Onlinesecs:
		//Idlesecs:
	}
	client.SendMessage(newUserStatsMsg)
}

// protocol handling dummy for UserRemoveMessage
func (messageHandler *MessageHandler) handleUserRemoveMessage(msg *Message) {
	msgProto := &mumbleproto.UserRemove{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("userRemoveMessage info:", msgProto, "from:", msg.client.UserName)
}

// protocol handling dummy for banlist message
func (m *MessageHandler) handleBanListMessage(msg *Message) {

	msgProto := &mumbleproto.BanList{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Banlist message info:", msgProto, "from:", msg.client.UserName)
}

//TODO : deal with sending text message
func (m *MessageHandler) handleTextMessage(msg *Message) {
	textMsg := &mumbleproto.TextMessage{}
	err := proto.Unmarshal(msg.buf, textMsg)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("Text msg info:", textMsg, "from:", msg.client.UserName)

	if textMsg.ChannelId != nil {
		newMsg := mumbleproto.TextMessage{
			Actor:     textMsg.Actor,
			Session:   textMsg.Session,
			ChannelId: textMsg.ChannelId,
			Message:   textMsg.Message,
		}

		servermodule.Cast(APIkeys.BroadcastChannel, newMsg, int(textMsg.ChannelId[0]))

	} else if textMsg.Session != nil {

	}

}

// protocol handling dummy for Aclmessage

func (m *MessageHandler) handleAclMessage(msg *Message) {
	msgProto := &mumbleproto.ACL{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ACL info:", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for QueryUSers

func (m *MessageHandler) handleQueryUsers(msg *Message) {
	msgProto := &mumbleproto.QueryUsers{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Handle query users  :", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for CtyptSetup
func (m *MessageHandler) handleCryptSetup(msg *Message) {
	msgProto := &mumbleproto.CryptSetup{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("handle cryptsetup :", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for UserList

func (m *MessageHandler) handleUserList(msg *Message) {
	msgProto := &mumbleproto.UserList{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(" userlist :", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for Voicetarget

func (m *MessageHandler) handleVoiceTarget(msg *Message) {
	msgProto := &mumbleproto.VoiceTarget{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("voice target info:", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for PermissionQuery
func (m *MessageHandler) handlePermissionQuery(msg *Message) {
	msgProto := &mumbleproto.PermissionQuery{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("permission query info:", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for requestBlob
func (m *MessageHandler) handleRequestBlob(msg *Message) {
	msgProto := &mumbleproto.RequestBlob{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("requestblob info:", msgProto, "from:", msg.client.UserName)

}

func (m *MessageHandler) handleChannelRemoveMessage(msg *Message) {
	msgProto := &mumbleproto.RequestBlob{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("channel remove info:", msgProto, "from:", msg.client.UserName)
}

// TODO : permission 처리 나누어서 구현
// Send message when permission denied
func sendPermissionDenied(client *Client, denyType mumbleproto.PermissionDenied_DenyType) {
	permissionDeniedMsg := &mumbleproto.PermissionDenied{
		Session: proto.Uint32(client.Session()),
		Type:    &denyType,
	}
	fmt.Println("Permission denied ", permissionDeniedMsg)
	err := client.SendMessage(permissionDeniedMsg)
	if err != nil {
		fmt.Println("Error sending messsage")
		return
	}

}

func (m *MessageHandler) Init() {
	servermodule.RegisterAPI((*MessageHandler).HandleMassage, APIkeys.HandleMessage)

}
