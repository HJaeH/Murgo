package data

import (
	"murgo/server"
)

type Message struct {
	buf    []byte
	kind   uint16
	client *server.TlsClient
	testCounter int
}


func (message *Message)SetBuf(buf []byte){
	message.buf = buf
}

func (message *Message)SetKind(kind uint16){
	message.kind = kind
}
func (message *Message)SetClient(client *server.TlsClient){
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

func (message *Message)Client ()(*server.TlsClient){
	return message.client
}

func (message *Message)TestCounter ()(int){
	return message.testCounter
}
