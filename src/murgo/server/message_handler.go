// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo 메시지 핸들러


package server


import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"mumble.info/grumble/pkg/mumbleproto"
)

type MessageHandler struct {

	supervisor *Supervisor
	Cast chan interface{} // todo 동기, 비동기 요청 두가지 채널로 구분
	Call chan interface{}


}


func NewMessageHandler(supervisor *Supervisor)( *MessageHandler){
	messageHandler := new(MessageHandler)
	messageHandler.supervisor = supervisor

	messageHandler.Cast = make( chan interface{})
	return messageHandler
}


func (messageHandler *MessageHandler) startMassageHandler(){
	fmt.Println("Message Handler stated")

	for{
		select {
		case castData := <-messageHandler.Cast:
			messageHandler.handleCast(castData)

		}
	}
}

func (messageHandler *MessageHandler) handleCast (castData interface{}) {
	//fmt.Println(" cast handler entered")

	switch t := castData.(type) {
	default:
		fmt.Printf("unexpected type %T", t)
	case *Message:
		msg := castData.(*Message)
		messageHandler.handleMassage(msg)

	}
}

func (messageHandler *MessageHandler) handleMassage(msg *Message) {

	//fmt.Println(" handler entered")
	switch msg.Kind() {
	case mumbleproto.MessageAuthenticate:
		messageHandler.handleAuthenticate(msg)

	case mumbleproto.MessagePing:
		messageHandler.handlePing(msg)
	//case mumbleproto.MessageChannelRemove:
	//	server.handleChannelRemoveMessage(msg.client, msg)
	case mumbleproto.MessageChannelState:
		messageHandler.handleChannelStateMessage(msg)
	case mumbleproto.MessageUserState:
		messageHandler.handleUserState(msg)
	//case mumbleproto.MessageUserRemove:
	//	server.handleUserRemoveMessage(msg.client, msg)
	//case mumbleproto.MessageBanList:
	//	server.handleBanListMessage(msg.client, msg)
	//case mumbleproto.MessageTextMessage:
	//	server.handleTextMessage(msg.client, msg)
	//case mumbleproto.MessageACL:
	//	server.handleAclMessage(msg.client, msg)
	//case mumbleproto.MessageQueryUsers:
	//	server.handleQueryUsers(msg.client, msg)
	//case mumbleproto.MessageCryptSetup:
	//	server.handleCryptSetup(msg.client, msg)
	//case mumbleproto.MessageContextAction:
	//	server.Printf("MessageContextAction from client")
	//case mumbleproto.MessageUserList:
	//	server.handleUserList(msg.client, msg)
	//case mumbleproto.MessageVoiceTarget:
	//	server.handleVoiceTarget(msg.client, msg)
	//case mumbleproto.MessagePermissionQuery:
	//	server.handlePermissionQuery(msg.client, msg)
	//case mumbleproto.MessageUserStats:
	//	server.handleUserStatsMessage(msg.client, msg)
	//case mumbleproto.MessageRequestBlob:
	//	server.handleRequestBlob(msg.client, msg)
	}
}

////Authenticate message handling
func (messageHandler *MessageHandler) handleAuthenticate(msg *Message) {
	//메시지 내용 출력 for test
	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.Buf(), authenticate)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Athenticate info:", authenticate)

	// crypto setup
	// Used to initialize and resync the UDP encryption. Either side may request a
	// resync by sending the message without any values filled. The resync is
	// performed by sending the message with only the client or server nonce
	// filled.
	client := msg.Client()
	client.cryptState.GenerateKey()
	err = client.sendMessage(&mumbleproto.CryptSetup{
		Key:         client.cryptState.Key(),
		ClientNonce: client.cryptState.EncryptIV(),
		ServerNonce: client.cryptState.DecryptIV(),
	})
	if err != nil{
		fmt.Println("error sending msg")
	}
	client.codecs = authenticate.CeltVersions
	if len(client.codecs) == 0 {
		//todo : no codec msg case
	}
	err = client.sendMessage(&mumbleproto.CodecVersion{
		Alpha:       proto.Int32(0),
		Beta:        proto.Int32(0),
		PreferAlpha: proto.Bool(true),
		Opus:        proto.Bool(true),
	})
	if err != nil {
		//server.Printf("Unable to broadcast.")
		return
	}
	/// send channel state
	channel := new(Channel)
	channel.Id = 1
	channel.Name = "myChannel"
	chanstate := &mumbleproto.ChannelState{
		ChannelId: proto.Uint32(uint32(channel.Id)),
		Name:      proto.String(channel.Name),
	}

	chanstate.Parent = proto.Uint32(uint32(10))
	chanstate.Description = proto.String(string("description"))

	/*if channel.IsTemporary() {
		chanstate.Temporary = proto.Bool(true)
	}*/
	chanstate.Temporary = proto.Bool(true)
	channel.Position = 0 //
	chanstate.Position = proto.Int32(int32(10))

	links := []uint32{}
	for cid, _ := range channel.Links {
		links = append(links, uint32(cid))
	}
	chanstate.Links = links

	err = client.sendMessage(chanstate)
	if err != nil {
		fmt.Println("error sending message")
		// todo
	}

	userstate := &mumbleproto.UserState{
		Session:   proto.Uint32(0),//client.Session()
		Name:      proto.String("user"), //client.ShownName()
		ChannelId: proto.Uint32(uint32(channel.Id)),
	}
	if err := client.sendMessage(userstate); err != nil {
		//client.Panicf("%v", err)
		return
	}
	sync := &mumbleproto.ServerSync{}
	sync.Session = proto.Uint32(0)
	sync.MaxBandwidth = proto.Uint32(1000000)
	sync.WelcomeText = proto.String("Welcome to Jaewha's murgo server")

	if err := client.sendMessage(sync); err != nil {
		fmt.Println("error sending message")
		return
	}

	err = client.sendMessage(&mumbleproto.ServerConfig{
		AllowHtml:          proto.Bool(true),
		MessageLength:      proto.Uint32(100),
		ImageMessageLength: proto.Uint32(100),
	})
	if err := client.sendMessage(sync); err != nil {
		fmt.Println("error sending message")
		return
	}

}



func (messageHandler *MessageHandler) handlePing(msg *Message) {
	ping :=&mumbleproto.Ping{}
	err := proto.Unmarshal(msg.Buf(), ping)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println("ping info:", ping, "msg id : ", msg.testCounter)

	client := msg.Client()
	client.sendMessage(&mumbleproto.Ping{
		Timestamp: ping.Timestamp,
		Good:      proto.Uint32(1),
		Late:      proto.Uint32(1),
		Lost:      proto.Uint32(1),
		Resync:    proto.Uint32(1),
	})
}



func (messageHandler *MessageHandler) handleChannelStateMessage(msg *Message) {

}

func (messageHandler *MessageHandler) handleUserState(msg *Message) {
	// 메시지를 보낸 유저 reset idle -> 이 부분은 통합
	userstate := &mumbleproto.UserState{}
	//messageHandler.supervisor.tc[1].handleCast()
	err := proto.Unmarshal(msg.buf, userstate)
	if err != nil {
		//
		return
	}
	clients := messageHandler.supervisor.tc
	channelManager := messageHandler.supervisor.cm

	actor, ok := clients[msg.Client().Session()]
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
			//client.Panic("Invalid session in UserState message")
			return
		}
	}

	userstate.Session = proto.Uint32(target.Session())
	userstate.Actor = proto.Uint32(actor.Session())


	//Channel ID 필드 값이 있는 경우
	if userstate.ChannelId != nil {
		channel, ok := channelManager.channelList[int(*userstate.ChannelId)]
		if ok {
			channelManager.Cast <- &MurgoMessage{kind:enterChannel, channel:channel }

			//broadcast = true
		}
	}

}


