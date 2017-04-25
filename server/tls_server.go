
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

func (tlsServer *TlsServer) startTlsServer() (err error) {
	fmt.Println("TlsServer stared")
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
	defer ln.Close()

	//accept loop와 cast handling 수행
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(" Accepting a conneciton failed handling a client")
				//continue
			}
			tlsServer.handleIncomingClient(conn)
		}
	}()


	return err
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

	//create client message receive loop as gen server
	// TODO : the start time need to be pushed back - after duplicate check
	// TODO : but the work is conducted in authenticate which is running in message accepting loop
	tlsServer.supervisor.startGenServer(tlsClient.recvLoop)
}

