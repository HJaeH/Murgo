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




func (message *Message)SetBuf(buf []byte){
	message.buf = buf
}

func (message *Message)SetKind(kind uint16){
	message.kind = kind
}
func (message *Message)SetClient(client *TlsClient){
	message.client = client
}
func (message *Message)SetTestCounter(testCounter int){
	message.testCounter = testCounter
}


func (message *Message)Buf ()([]byte){
	return message.buf
}

func (message *Message)Kind ()(uint16){
	return message.kind
}

func (message *Message)Client ()(*TlsClient){
	return message.client
}

func (message *Message)TestCounter ()(int){
	return message.testCounter
}
