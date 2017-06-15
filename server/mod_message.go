package server

import (
	"fmt"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	APIkeys "murgo/server/util"

	"github.com/golang/protobuf/proto"
)

type MessageHandler struct{}

type Message struct {
	buf    []byte
	kind   uint16
	client *Client
}

func (messageHandler *MessageHandler) HandleMessage(msg *Message) {
	switch msg.kind {

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
	//case mumbleproto.MessageContextAction:
	//	fmt.Print("MessageContextAction from client")
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
	default:
		fmt.Println("uncategorized msg type :", msg.kind)
	}
}

func (m *MessageHandler) handlePingMessage(msg *Message) {
	ping := &mumbleproto.Ping{}
	err := proto.Unmarshal(msg.buf, ping)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := msg.client
	fmt.Println(ping)
	client.sendMessage(&mumbleproto.Ping{
		Timestamp:  ping.Timestamp,
		TcpPackets: ping.TcpPackets,
		TcpPingVar: ping.TcpPingVar,
		TcpPingAvg: ping.TcpPingAvg,
	})

}

func (m *MessageHandler) handleUserStateMessage(msg *Message) {

	// 메시지를 보낸 유저 reset idle -> 이 부분은 통합
	userState := &mumbleproto.UserState{}
	err := proto.Unmarshal(msg.buf, userState)
	if err != nil {
		panic("error while unmarshalling")
		return
	}
	//Channel ID 필드 값이 있는 경우
	if userState.ChannelId != nil {
		servermodule.Call(APIkeys.EnterChannel, *userState.ChannelId, msg.client)
	}
	//servermodule.Cast(APIkeys.SetUserOption, msg.client, userState)

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
	//fmt.Println("userstats info:", userStats, "from:", msg.client.UserName)
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
	client := msg.client
	textMsg := &mumbleproto.TextMessage{}
	err := proto.Unmarshal(msg.buf, textMsg)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("Text msg info:", textMsg, "from:", msg.client.UserName)

	//todo : 예외처리
	/*if len(textMsg.Message) == 0 {
		return
	}*/
	newMsg := &mumbleproto.TextMessage{
		Actor:   proto.Uint32(client.Session()),
		Message: textMsg.Message,
	}
	// send text message to channels
	for _, eachChannelId := range textMsg.ChannelId {
		servermodule.AsyncCall(APIkeys.BroadCastChannelWithoutMe, eachChannelId, client, newMsg)
	}

	// send text message to users
	servermodule.AsyncCall(APIkeys.SendMessages, textMsg.Session, newMsg)
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
	err := client.sendMessage(permissionDeniedMsg)
	if err != nil {
		fmt.Println("Error sending messsage")
		return
	}

}

func (m *MessageHandler) Init() {
	servermodule.RegisterAPI((*MessageHandler).HandleMessage, APIkeys.HandleMessage)

}
