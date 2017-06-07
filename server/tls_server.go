package server

import (
	"fmt"

	"crypto/tls"
	"io"
	"murgo/config"
	"murgo/pkg/servermodule"
	APIkeys "murgo/server/util"
	"net"
)

type TlsServer struct {
	ln net.Listener

	//todo : need to be deleted
}

func (tlsServer *TlsServer) Accept() {
	conn, err := tlsServer.ln.Accept()
	if err != nil {
		fmt.Println(" Accepting a conneciton failed handling a client")
	}
	servermodule.Cast(APIkeys.HandleIncomingClient, conn)
	servermodule.Cast(APIkeys.Accept)

}

func (tlsServer *TlsServer) Init() {
	servermodule.RegisterAPI((*TlsServer).Receive, APIkeys.Receive)
	servermodule.RegisterAPI((*TlsServer).Accept, APIkeys.Accept)
	cer, err := tls.LoadX509KeyPair("./src/murgo/config/server.crt", "./src/murgo/config/server.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	//server start to listen on tls
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	tlsServer.ln, err = tls.Listen(config.CONN_TYPE, config.DEFAULT_PORT, tlsConfig)

	/*ln, err := net.Listen(config.CONN_TYPE, config.DEFAULT_PORT)
	tlsServer.ln = ln*/
	if err != nil {
		fmt.Println(err)
		return
	}
	//todo : accept routine into framework
	// todo : make accept as a cast req

	servermodule.Cast(APIkeys.Accept)
}

/*
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
}*/

func (ts *TlsServer) terminate() {
	ts.ln.Close()
	//todo : 메모리 회수 및 나머지 작업

}

func (t *TlsServer) Receive(client *Client) {
	// todo : goroutine is to be blocked here,
	// todo : check the trade offs of keeping one io reader and this
	for {
		msg, err := client.readProtoMessage()
		if err != nil {
			if err != nil {
				if err == io.EOF {
					client.Disconnect()
				} else {
					//client.Panicf("%v", err)
				}
				return
			}
		}
		servermodule.Cast(APIkeys.HandleMessage, msg)
		/*servermodule.Cast(APIkeys.Receive, client)*/
	}
}
