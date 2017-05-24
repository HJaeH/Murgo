// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// server tls accept server

package server


import (
	"fmt"
	"crypto/tls"

	"murgo/config"
	"murgo/pkg/sessionpool"
)




func NewTlsServer(supervisor *MurgoSupervisor) (*TlsServer) {
	tlsServer := new(TlsServer)
	tlsServer.supervisor = supervisor

	return tlsServer

}

func startTlsServer() {
	fmt.Println("TlsServer stared")
	// tls setting
	cer, err := tls.LoadX509KeyPair("./src/murgo/config/server.crt", "./src/murgo/config/server.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	//server start to listen on tls
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen(config.CONN_TYPE, config.DEFAULT_PORT, tlsConfig)
	defer ln.Close()
	if err != nil {

		fmt.Println(err)
		return
	}

	//accept loop
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(" Accepting a conneciton failed handling a client")
			//continue
		}
		tlsServer.supervisor.sessionManager.Cast <- &MurgoMessage{
			Kind:handleIncomingClient,
			Conn:&conn,
		}

	}
}

