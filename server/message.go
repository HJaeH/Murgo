package server

import "net"

type Message struct {
	buf    []byte
	kind   uint16
	client *TlsClient
	testCounter int
}

type MurgoMessage struct {
	kind uint16
	channelId int
	client *TlsClient
	channel *Channel
	msg interface{}
	ChannelName string
	conn *net.Conn

}

