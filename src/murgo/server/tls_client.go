
package server


import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/golang/protobuf/proto"
	"net"
	"io"
	"mumble.info/grumble/pkg/mumbleproto"
	"bufio"
	"murgo/config"
)


type TlsClient struct {
	supervisor *Supervisor
	cast chan interface{}



	conn net.Conn
	session uint32


	pid  int64

	username string
	server *TlsServer
	reader  *bufio.Reader

	tcpaddr *net.TCPAddr
	certHash string



	//client auth infomations
	codecs []int32
	tokens []string

	//crypt state
	cryptState *config.CryptState




	//
	testCounter int
}


func NewTlsClient(supervisor *Supervisor, conn net.Conn) (*TlsClient){

	//create new object
	tlsClient := new(TlsClient)
	tlsClient.cryptState = new(config.CryptState)

	//set servers
	tlsClient.supervisor = supervisor
	tlsClient.server = supervisor.ts

	//
	tlsClient.conn = conn
	tlsClient.session = tlsClient.server.sessionPool.Get()
	tlsClient.reader = bufio.NewReader(tlsClient.conn)

	tlsClient.testCounter = 0
	return tlsClient
}

func (tlsClient *TlsClient) startTlsClient(){

	go tlsClient.recvLoop()
	for {

		select {
		case castData := <-tlsClient.cast:
			tlsClient.castHandler(castData)
		}
	}
}

func (tlsClient *TlsClient) castHandler (castData interface{}) {

	//tlsClient.readProtoMessage()
}




//send msg to client
func (tlsClient *TlsClient) sendMessage(msg interface{}) error {
	buf := new(bytes.Buffer)
	var (
		kind    uint16
		msgData []byte
		err     error
	)

	kind = mumbleproto.MessageType(msg)
	if kind == mumbleproto.MessageUDPTunnel {
		msgData = msg.([]byte)
	} else {
		protoMsg, ok := (msg).(proto.Message)
		if !ok {
			return errors.New("client: exepcted a proto.Message")
		}
		msgData, err = proto.Marshal(protoMsg)
		if err != nil {
			return err
		}
	}

	err = binary.Write(buf, binary.BigEndian, kind)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, uint32(len(msgData)))
	if err != nil {
		return err
	}
	_, err = buf.Write(msgData)
	if err != nil {
		return err
	}

	_, err = tlsClient.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

//
func (tlsClient *TlsClient) readProtoMessage() (msg *Message, err error) {
	var (
		length uint32
		kind   uint16
	)

	// Read the message type (16-bit big-endian unsigned integer)
	//read data form io.reader
	err = binary.Read(tlsClient.reader, binary.BigEndian, &kind)
	if err != nil {
		return
	}

	// Read the message length (32-bit big-endian unsigned integer)
	err = binary.Read(tlsClient.reader, binary.BigEndian, &length)
	if err != nil {
		return
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(tlsClient.reader, buf)
	if err != nil {
		return
	}
	tlsClient.testCounter++
	msg = &Message{
		buf:    buf,
		kind:   kind,
		client: tlsClient,
		testCounter: tlsClient.testCounter,
	}

	return msg, err
}

func (tlsClient *TlsClient) recvLoop (){

	for {
		msg, err := tlsClient.readProtoMessage()
		if err != nil {
			if err != nil {
				if err == io.EOF {
					tlsClient.Disconnect()
				} else {
					//client.Panicf("%v", err)
				}
				return
			}
		}
		tlsClient.supervisor.mh.cast <- msg
	}
}


//TODO
func (tlsClient *TlsClient) Disconnect() {


}



