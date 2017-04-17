// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo 메시지 핸들러


package server


import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"mumble.info/grumble/pkg/mumbleproto"
	"mumble.info/grumble/pkg/acl"
)

type MessageHandler struct {

	supervisor *Supervisor
	cast chan interface{}

}
type Message struct {
	buf    []byte
	kind   uint16
	client *TlsClient
	testCounter int
}

func NewMessageHandler(supervisor *Supervisor)( *MessageHandler){
	messageHandler := new(MessageHandler)
	messageHandler.supervisor = supervisor

	messageHandler.cast = make( chan interface{})
	return messageHandler
}


func (messageHandler *MessageHandler) startMassageHandler(){
	for{
		select {
		case castData := <-messageHandler.cast:
			messageHandler.castHandler(castData)

		}
	}
}

func (messageHandler *MessageHandler) castHandler (castData interface{}) {
	//fmt.Println(" casthandler entered")

	switch t := castData.(type) {
	default:
		fmt.Printf("unexpected type %T", t)
	case *Message:
		msg := castData.(*Message)
		msg.MessageHandler()
	}
}




func (msg *Message) MessageHandler() {
	//fmt.Println(" handler entered")
	switch msg.kind {
	case mumbleproto.MessageAuthenticate:
		msg.handleAuthenticate()

	case mumbleproto.MessagePing:
		msg.handlePing()
	//case mumbleproto.MessageChannelRemove:
	//	server.handleChannelRemoveMessage(msg.client, msg)
	//case mumbleproto.MessageChannelState:
	//	server.handleChannelStateMessage(msg.client, msg)
	case mumbleproto.MessageUserState:
		msg.handleUserState()
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
func (msg *Message) handleAuthenticate() {
	//메시지 내용 출력 for test
	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, authenticate)
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
	client := msg.client
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


func (msg *Message)handleUserState()  {
	fmt.Println("000000000000")
	userState := &mumbleproto.UserState{}
	err := proto.Unmarshal(msg.buf, userState)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Userstate info:", userState ,"msg id : ", msg.testCounter)
}


func (msg *Message) handlePing() {
	ping :=&mumbleproto.Ping{}
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




