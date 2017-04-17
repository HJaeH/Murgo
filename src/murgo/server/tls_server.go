
package server


import (
	"fmt"
	"net"
	"crypto/tls"
	"log"
	"github.com/golang/protobuf/proto"

	"mumble.info/grumble/pkg/mumbleproto"
	"murgo/config"
	"mumble.info/grumble/pkg/sessionpool"
	"os"
)

type TlsServer struct {
	supervisor *Supervisor
	sessionPool *sessionpool.SessionPool
	tlsConfig  *tls.Config
	cast chan interface{}
}
/*

type Server struct {
	SID int64			//Server ID
	incoming chan(*Message)
	tlsConfig  *tls.Config



	//usernames map[string]*User
	//users  map[uint32]*User


}*/

func NewTlsServer(supervisor *Supervisor) (*TlsServer) {
	tlsServer := new(TlsServer)
	tlsServer.supervisor = supervisor
	tlsServer.sessionPool = sessionpool.New()

	return tlsServer

}

func (tlsServer *TlsServer) startTlsServer() (err error){

	gopath := os.Getenv("GOPATH")

	fmt.Println("Launching server...")
	// tls setting
	cer, err := tls.LoadX509KeyPair(gopath+"/src/murgo/config/server.crt", gopath+"/src/murgo/config/server.key")
	if err != nil {
		log.Println(err)
		return
	}
	//server start to listen on tls
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen(config.CONN_TYPE, config.DEFAULT_PORT, tlsConfig)
	if err != nil {
		log.Println(err)
		return
	}
	//defer ln.Close()

	//accept loop와 cast handling 수행
	//connChannel := make(chan net.Conn)
	go func(){
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(" Accepting a conneciton failed handling a client")
			}
			tlsServer.handleIncomingClient(conn)
		}
	}()

	for {
		select {
		case castData := <-tlsServer.cast:
			tlsServer.castHandler(castData)
		}
		//accept owns the flow until new client connected
	}
}


func (tlsServer *TlsServer)handleIncomingClient (conn net.Conn){
	fmt.Println("test2")

	//init tls client
	tlsClient := NewTlsClient(tlsServer.supervisor, conn)

	if tlsServer.supervisor.tc[tlsClient.session] != nil {
		// todo
	}
	tlsServer.supervisor.tc[tlsClient.session] = tlsClient

	// send version information
	version := &mumbleproto.Version{
		Version:     proto.Uint32(0x10205),
		Release:     proto.String("Murgo"),
		CryptoModes: config.SupportedModes(),
	}
	err := tlsClient.sendMessage(version)
	if err != nil {
		fmt.Println("Error sending message to client")
	}
	fmt.Println("test2")

	//supervisor에서 클라이언트 고루틴 생성
	tlsServer.supervisor.cast <- tlsClient.session
	fmt.Println("test3")
}




func (tlsServer *TlsServer) castHandler (castData interface{}) {

	switch t := castData.(type) {
	default:
		fmt.Printf(" : unexpected type %T", t)

	case uint32:
		//Todo
	}
}