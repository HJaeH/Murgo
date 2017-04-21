
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
)

type TlsServer struct {
	supervisor *Supervisor
	sessionPool *sessionpool.SessionPool
	tlsConfig  *tls.Config
	Cast chan interface{}
	Call chan interface{}


}


func NewTlsServer(supervisor *Supervisor) (*TlsServer) {
	tlsServer := new(TlsServer)
	tlsServer.supervisor = supervisor
	tlsServer.sessionPool = sessionpool.New()

	return tlsServer

}

func (tlsServer *TlsServer) startTlsServer() (err error){
	fmt.Println("TlsServer stated")
	// tls setting
	cer, err := tls.LoadX509KeyPair("./src/murgo/config/server.crt", "./src/murgo/config/server.key")

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
		case castData := <-tlsServer.Cast:

			tlsServer.handleCast(castData)
		}
		//accept owns the flow until new client connected
	}
}


func (tlsServer *TlsServer)handleIncomingClient (conn net.Conn){

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

	//supervisor에서 클라이언트 고루틴 생성
	//tlsServer.supervisor.Cast <- tlsClient.session
	tlsServer.supervisor.startGenServer(tlsClient.recvLoop)

}




func (tlsServer *TlsServer) handleCast(castData interface{}) {

	switch t := castData.(type) {
	default:
		fmt.Printf(" : unexpected type %T", t)

	case uint32:
	//Todo
	}
}

