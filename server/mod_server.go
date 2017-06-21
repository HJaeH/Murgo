package server

import (
	"fmt"

	"crypto/tls"
	"io"
	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"

	"murgo/server/util/apikeys"
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
		return
	}
	servermodule.AsyncCall(apikeys.HandleIncomingClient, conn)
	servermodule.AsyncCall(apikeys.Accept)
}

func (s *Server) Init() {
	servermodule.RegisterAPI((*Server).Receive, apikeys.Receive)
	servermodule.RegisterAPI((*Server).Accept, apikeys.Accept)
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

	servermodule.AsyncCall(apikeys.Accept)
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
					//log.Panic("Client left")
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
			//Do not send voice data to clients in root channel
			if client.Channel.Id == ROOT_CHANNEL {
				continue
			} else {
				buf := msg.buf
				if len(buf) == 0 {
					return
				}
				kind := (msg.buf[0] >> 5) & 0x07
				switch kind {
				case mumbleproto.UDPMessageVoiceOpus:
					client.addFrame(uint32(len(msg.buf)))
					outbuf := make([]byte, 1024)
					incoming := packetdata.New(buf[1 : 1+(len(buf)-1)])
					outgoing := packetdata.New(outbuf[1 : 1+(len(outbuf)-1)])
					_ = incoming.GetUint32()
					size := int(incoming.GetUint16())
					incoming.Skip(size & 0x1fff)
					outgoing.PutUint32(client.Session())
					outgoing.PutBytes(buf[1 : 1+(len(buf)-1)])
					outbuf[0] = buf[0] & 0xe0
					buf := outbuf[0 : 1+outgoing.Size()]
					go client.Channel.BroadCastChannelWithoutMe(buf, client)
				}
			}
		} else {
			servermodule.AsyncCall(apikeys.HandleMessage, msg)
		}
	}
}
