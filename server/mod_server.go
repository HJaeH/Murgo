package server

import (
	"fmt"

	"crypto/tls"
	"io"
	"murgo/config"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"

	"net"

	"murgo/server/util/event"

	"time"

	"murgo/pkg/servermodule/log"

	"mumble.info/grumble/pkg/packetdata"
)

type Server struct {
	ln net.Listener
}

func (s *Server) Accept() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Error(" Accepting a conneciton failed handling a client")
			continue
		}
		servermodule.AsyncCall(event.HandleIncomingClient, conn)
	}
}

func (s *Server) Init() {
	servermodule.EventRegister((*Server).Receive, event.Receive)
	servermodule.EventRegister((*Server).Accept, event.Accept)
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

	servermodule.AsyncCall(event.Accept)
}

func (s *Server) terminate() {
	s.ln.Close()
}

func (s *Server) Receive(client *Client) error {

	for {
		msg, err := client.readProtoMessage()
		client.timer.Reset(time.Second * 10)
		if err != nil {
			if err != nil {
				if err == io.EOF {
					//client disconnected
					client.Disconnect()
					return err
				} else {
					return err
				}
			}
		}

		if msg.kind == mumbleproto.MessageUDPTunnel {
			//Do not send voice data to clients in root channel
			if client.Channel == nil || client.Channel.Id == ROOT_CHANNEL || client.mute || client.selfMute {
				continue
			} else {
				buf := msg.buf
				if len(buf) == 0 {
					return nil
				}
				kind := (msg.buf[0] >> 5) & 0x07
				switch kind {
				case mumbleproto.UDPMessageVoiceOpus:
					client.addFrame(uint32(len(msg.buf)))
					buf := parseVoiceMessage(client, buf)
					go client.Channel.broadcastVoice(buf, client)
				}
			}
		} else {
			servermodule.AsyncCall(event.HandleMessage, msg)
		}

		select {
		//client timeout
		case <-client.timer.C:
			client.Disconnect()
			return nil
		default:

		}
	}
	return nil
}

func parseVoiceMessage(client *Client, buf []byte) []byte {
	outbuf := make([]byte, 1024)
	incoming := packetdata.New(buf[1 : 1+(len(buf)-1)])
	outgoing := packetdata.New(outbuf[1 : 1+(len(outbuf)-1)])
	size := int(incoming.GetUint16())

	incoming.Skip(size & 0x1fff)
	outgoing.PutUint32(client.Session())
	outgoing.PutBytes(buf[1 : 1+(len(buf)-1)])
	outbuf[0] = buf[0] & 0xe0

	return outbuf[0 : 1+outgoing.Size()]

}
