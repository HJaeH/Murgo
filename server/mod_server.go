package server

import (
	"fmt"

	"crypto/tls"
	"io"
	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"

	APIkeys "murgo/server/util"
	"net"

	"mumble.info/grumble/pkg/packetdata"
)

type Server struct {
	ln net.Listener
}

func (s *Server) Accept() {
	conn, err := s.ln.Accept()
	if err != nil {
		fmt.Println(" Accepting a conneciton failed handling a client")
	}
	servermodule.AsyncCall(APIkeys.HandleIncomingClient, conn)
	servermodule.AsyncCall(APIkeys.Accept)

}

func (s *Server) Init() {
	servermodule.RegisterAPI((*Server).Receive, APIkeys.Receive)
	servermodule.RegisterAPI((*Server).Accept, APIkeys.Accept)
	cer, err := tls.LoadX509KeyPair("./src/murgo/config/server.crt", "./src/murgo/config/server.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	//server start to listen on tls
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
	s.ln, err = tls.Listen(config.CONN_TYPE, config.DEFAULT_PORT, tlsConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	servermodule.AsyncCall(APIkeys.Accept)
}

func (s *Server) terminate() {
	s.ln.Close()

}

func (s *Server) Receive(client *Client) {

	for {

		msg, err := client.readProtoMessage()
		if err != nil {
			if err != nil {
				if err == io.EOF {
					client.Disconnect()
				} else {
					//client disconnected
					return
					//panic(err)
				}
				return
			}
		}
		if msg.kind == mumbleproto.MessageUDPTunnel {

			//VoicelibTester client also check this, just to be sure.
			if client.Channel.Id == ROOT_CHANNEL {
				continue
			} else {
				buf := msg.buf
				if len(buf) == 0 {
					return
				}

				kind := (msg.buf[0] >> 5) & 0x07
				//fmt.Print(msg.buf, ", ")
				switch kind {

				case mumbleproto.UDPMessageVoiceOpus:
					outbuf := make([]byte, 1024)

					incoming := packetdata.New(buf[1 : 1+(len(buf)-1)])
					outgoing := packetdata.New(outbuf[1 : 1+(len(outbuf)-1)])
					_ = incoming.GetUint32()

					size := int(incoming.GetUint16())
					incoming.Skip(size & 0x1fff)
					outgoing.PutUint32(client.Session())
					outgoing.PutBytes(buf[1 : 1+(len(buf)-1)])
					outbuf[0] = buf[0] & 0xe0 // strip target

					go client.Channel.BroadCastChannelWithoutMe(client, buf)
					//servermodule.Cast(APIkeys.BroadCastVoiceToChannel, client, outbuf)
				}

			}
		} else {
			servermodule.AsyncCall(APIkeys.HandleMessage, msg)
		}
	}
}
