package main

import (
	"fmt"
	"net"
	"crypto/tls"
	"log"
	"../pkg/protobuf"
	"github.com/golang/protobuf/proto"
	"crypto/sha1"
	"encoding/hex"
	"io"
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

func  AddServer(id int64) (s *Server, err error) {

	s = new(Server)
	id = s.SID


	return s, err
}


func (server *Server) Start() (err error){

	fmt.Println("Launching server...")
	//server := new(Server)
	//server.tcpl, err = net.ListenTCP(CONN_TYPE, &net.TCPAddr{IP: net.ParseIP(CONN_HOST), Port: CONN_PORT})
	/*if err != nil {
		return err
	}*/

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

	fmt.Println("check point 1")
	for {
		//accept owns the flow until new client connected
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(" Accepting a conneciton failed handling a client")
			continue
		}
		go server.HandleClientConnection(conn)

	}


}
func (server *Server) HandleClientConnection(conn net.Conn){

	stringArray := []string {"","",""}
	_ = stringArray
	temp := make([]byte, 128)
	_ = temp


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
	client.sendMessage(version, conn)

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

		//version := &mumbleproto.Version{}
		fmt.Println("msg : ", msg.kind)

		//server.incoming<- msg

		/*err = proto.Unmarshal(msg.buf, version)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print("Message Received:", version)
*/

		/*
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message Received:", string(message))
		//fmt.Print("Message Received:")
		val, _ := client.conn.Read(temp)
		// go client.readPacket(temp)

		fmt.Println(val," ::::",temp, " - received");
		*/
	}


	// output message received



	// sample process for string received
	//newmessage := strings.ToUpper(message)
	// send new string back to client
	//conn.Write([]byte(newmessage + "\n"))

}
/*


func (server *Server) handleAuthenticate(client *Client, msg *Message) {
	// Is this message not an authenticate message? If not, discard it...
	if msg.kind != mumbleproto.MessageAuthenticate {
		//client.Panic("Unexpected message. Expected Authenticate.")
		return
	}

	auth := &mumbleproto.Authenticate{}
	err := proto.Unmarshal(msg.buf, auth)
	if err != nil {
		//client.Panic("Unable to unmarshal Authenticate message.")
		return
	}


	// Did we get a username?
	if auth.Username == nil || len(*auth.Username) == 0 {
		//client.RejectAuth(mumbleproto.Reject_InvalidUsername, "Please specify a username to log in")

		return
	}

	client.username = *auth.Username


	// First look up registration by name.
	user, exists := server.UserNameMap[client.username]
	if exists {
		if client.HasCertificate() && user.CertHash == client.CertHash() {
			client.user = user
		} else {
			client.RejectAuth(mumbleproto.Reject_WrongUserPW, "Wrong certificate hash")
			return
		}
	}

	// Name matching didn't do.  Try matching by certificate.
	if client.user == nil && client.HasCertificate() {
		user, exists := server.UserCertMap[client.CertHash()]
		if exists {
			client.user = user
		}
	}


	// Setup the cryptstate for the client.
	err = client.crypt.GenerateKey(client.CryptoMode)
	if err != nil {
		client.Panicf("%v", err)
		return
	}

	// Send CryptState information to the client so it can establish an UDP connection,
	// if it wishes.
	client.lastResync = time.Now().Unix()
	err = client.sendMessage(&mumbleproto.CryptSetup{
		Key:         client.crypt.Key,
		ClientNonce: client.crypt.DecryptIV,
		ServerNonce: client.crypt.EncryptIV,
	})
	if err != nil {
		client.Panicf("%v", err)
	}

	// Add codecs
	client.codecs = auth.CeltVersions
	client.opus = auth.GetOpus()

	client.state = StateClientAuthenticated
	server.clientAuthenticated <- client
}
*/
