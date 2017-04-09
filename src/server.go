package main

import (
	"fmt"
	"net"
	"crypto/tls"
	"log"
	"../pkg/protobuf"
	"github.com/golang/protobuf/proto"
	"io"
	"crypto/sha1"
	"encoding/hex"
)

type Server struct {
	SID int64
	incoming chan(*Message)

	tcpl    *net.TCPListener
	tlsl    net.Listener
	//udpconn *net.UDPConn
	tlscfg  *tls.Config


	//usernames map[string]*User
	//users  map[uint32]*User


}

func  CreateServer(id int64) (s *Server, err error) {

	s = new(Server)
	s.SID = id
	s.incoming = make (chan *Message)

	return s, err
}


func (server *Server) StartServer() (err error){

	fmt.Println("Launching server...")

	// tls setting
	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Println(err)
		return
	}

	//server start to listen on tls
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, _ := tls.Listen(CONN_TYPE, DEFAULT_PORT, config)
	defer ln.Close()
	//ln, _ := net.Listen(CONN_TYPE, CONN_PORT)
	//a := make (chan int)
	go server.MessageReceiver()
	i := 0
	for {
		//accept owns the flow until new client connected
		conn, err := ln.Accept()
		i=i+1
		//a <- i
		if err != nil {
			fmt.Println(" Accepting a conneciton failed handling a client")
			continue
		}
		go server.HandleClientConnection(conn)

	}


}
func (server *Server) HandleClientConnection(conn net.Conn){

	//stringArray := []string {"","",""}
	//_ = stringArray
	//temp := make([]byte, 128)
	//_ = temp


	client := new(Client)
	client.init(conn)


	tlsconn := conn.(*tls.Conn)
	err := tlsconn.Handshake()
	if err != nil {
		//client.Printf("TLS handshake failed: %v", err)
		fmt.Println("TLS handshake failed: %v", err)
		client.Disconnect()
	}


	state := tlsconn.ConnectionState()
	if len(state.PeerCertificates) > 0 {
		hash := sha1.New()
		hash.Write(state.PeerCertificates[0].Raw)
		sum := hash.Sum(nil)
		client.certHash = hex.EncodeToString(sum)
	}

	version := &mumbleproto.Version{
		Version:     proto.Uint32(0x10205),
		Release:     proto.String("Murgo"),
		CryptoModes: SupportedModes(),
	}
	err = client.sendMessage(version)
	if err != nil {
		//todo
	}


	client.cryptState.GenerateKey()
	err = client.sendMessage(&mumbleproto.CryptSetup{
		Key:         client.cryptState.key,
		ClientNonce: client.cryptState.encryptIV,
		ServerNonce: client.cryptState.decryptIV,
	})
	if err != nil{
		fmt.Println("error sending msg")
	}




	for {
		msg, err := client.readProtoMessage() // server ver
		if err != nil {
			if err == io.EOF {
				client.Disconnect()
			} else {
				//client.Panicf("%v", err)
			}
			return
		}
		fmt.Println("received message type : ", msg.kind)
		server.incoming<- msg
	}
}


func (server *Server) ReceivedMessageHandler (msg *Message) {
	switch msg.kind {
	case mumbleproto.MessageAuthenticate:
		server.HandleAuthenticate(msg)

	//case mumbleproto.MessagePing:
	//	server.handlePingMessage(msg.client, msg)
	//case mumbleproto.MessageChannelRemove:
	//	server.handleChannelRemoveMessage(msg.client, msg)
	//case mumbleproto.MessageChannelState:
	//	server.handleChannelStateMessage(msg.client, msg)
	case mumbleproto.MessageUserState:
		server.HandleUserStateMessage(msg)
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

func (server *Server) MessageReceiver() {
	fmt.Println("message receiver running")
	for{
		select {
		case msg := <-server.incoming:
			server.ReceivedMessageHandler(msg)
		//case temp := <- a:
		//	fmt.Println("Message receiver 2 ", temp)
		}
	}
}


////Authenticate message 를 받았 을 때 서버 작업
func (server *Server) HandleAuthenticate(msg *Message) {

	//메시지 내용 출력 for test
	authenticate := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, authenticate)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Athenticate info:", authenticate)
	//////

	client := msg.client
	client.cryptState.GenerateKey()
	err = client.sendMessage(&mumbleproto.CryptSetup{
		Key:         client.cryptState.key,
		ClientNonce: client.cryptState.encryptIV,
		ServerNonce: client.cryptState.decryptIV,
	})
	if err != nil{
		fmt.Println("error sending msg")
	}
	client.codecs = authenticate.CeltVersions
	if len(client.codecs) == 0 {
		//todo : no codec msg case
	}
	fmt.Println("check point")
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
	chanstate := &mumbleproto.ChannelState{
		ChannelId: proto.Uint32(uint32(channel.Id)),
		Name:      proto.String(channel.Name),
	}

	chanstate.Parent = proto.Uint32(uint32(10))
	chanstate.Description = proto.String(string("dd"))

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

func (server *Server)HandleUserStateMessage(msg *Message)  {
	userState := &mumbleproto.UserState{}
	err := proto.Unmarshal(msg.buf, userState)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Userstate info:", userState)
}