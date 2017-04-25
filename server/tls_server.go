
package server


import (
	"fmt"
	"crypto/tls"
	"log"

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

	return tlsServer

}

func (tlsServer *TlsServer) startTlsServer() {
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
	defer ln.Close()
	if err != nil {

		log.Println(err)
		return
	}

	//accept loop와 cast handling 수행
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(" Accepting a conneciton failed handling a client")
			//continue
		}
		tlsServer.supervisor.sm.Cast <- &MurgoMessage{
			kind:handleIncomingClient,
			conn:&conn,
		}

	}
}

