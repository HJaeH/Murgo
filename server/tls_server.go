// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// server tls accept server

package server

import (
	"crypto/tls"
	"fmt"

	"murgo/config"
	"murgo/pkg/servermodule"
	"net"
	"reflect"
)

const (
	accept = tlsserver + iota
)

type TlsServer struct {
	ln net.Listener

	//todo : need to be deleted
}

//const accept = servermodule.RegisterAPI(TlsServer.Accept)

func (tlsServer *TlsServer) Accept() {
	fmt.Println("server is listening")
	conn, err := tlsServer.ln.Accept()
	if err != nil {
		fmt.Println(" Accepting a conneciton failed handling a client")
		//continue
	}
	servermodule.Cast((*SessionManager).handleIncomingClient, &conn)
	servermodule.Cast((*TlsServer).Accept)

}

func (tlsServer *TlsServer) Init() {

	servermodule.RegisterAPI((*TlsServer).Accept)

	//
	cer, err := tls.LoadX509KeyPair("./src/murgo/config/server.crt", "./src/murgo/config/server.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	//server start to listen on tls
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	tlsServer.ln, err = tls.Listen(config.CONN_TYPE, config.DEFAULT_PORT, tlsConfig)
	if err != nil {

		fmt.Println(err)
		return
	}
	//todo : accept routine into framework
	// todo : make accept as a cast req

	servermodule.Cast(tlsserver, accept)

}

func (ts *TlsServer) HandleCast(request string, args ...interface{}) {

	if args == nil {
		reflect.ValueOf(ts).MethodByName(request).Call([]reflect.Value{})
	} else {
		inputs := make([]reflect.Value, len(args))
		for i, _ := range args {
			inputs[i] = reflect.ValueOf(args[i])
		}
		reflect.ValueOf(ts).MethodByName(request).Call(inputs)
	}
}

func (ts *TlsServer) HandleCall(request string, args ...interface{}) {
	if args == nil {
		reflect.ValueOf(ts).MethodByName(request).Call([]reflect.Value{})
	} else {
		inputs := make([]reflect.Value, len(args))
		for i, _ := range args {
			inputs[i] = reflect.ValueOf(args[i])
		}
		reflect.ValueOf(ts).MethodByName(request).Call(inputs)
	}
}

func (ts *TlsServer) terminate() {
	ts.ln.Close()
	//todo : 메모리 회수 및 나머지 작업

}
