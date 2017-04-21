
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
	"mumble.info/grumble/pkg/acl"
	"fmt"
)


type TlsClient struct {
	supervisor *Supervisor

	// 유저가 접속중인 channel
	channel *Channel


	conn net.Conn
	session uint32

	username string
	server *TlsServer
	reader  *bufio.Reader

	tcpaddr *net.TCPAddr
	certHash string


	//user's setting
	selfDeaf bool
	selfMute bool

	//client auth infomations
	codecs []int32
	tokens []string

	//crypt state
	cryptState *config.CryptState

	//for test
	testCounter int

}

// write 작업과 read 작업 구분 필요


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

	// 기본으로 루트채널에 할당
	tlsClient.channel = supervisor.cm.RootChannel()
	return tlsClient
}

func (tlsClient *TlsClient) recvLoop( ){
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
		tlsClient.supervisor.mh.Cast <- msg

	}
}



const (
	message uint16 = iota
)

func (tlsClient *TlsClient)handleCast( castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch murgoMsg.kind {
	default:
		fmt.Printf("unexpected type")
	case message:
		tlsClient.sendMessage(murgoMsg.msg)
	//todo
	}


}





///// internal functions

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
	msg = &Message{}
	/*{
		buf:    buf,
		kind:   kind,
		client: tlsClient,
		testCounter: tlsClient.testCounter,
	}*/
	msg.SetBuf(buf)
	msg.SetClient(tlsClient)
	msg.SetKind(kind)
	msg.SetTestCounter(tlsClient.testCounter)


	return msg, err
}


func (tlsClient *TlsClient) Disconnect() {

}

func (tlsClient *TlsClient) Session()(uint32) {
	return tlsClient.session
}



// Send permission denied by who, what, where
func (tlsClient *TlsClient)sendPermissionDenied(who *TlsClient, where *Channel, what acl.Permission) {
	pd := &mumbleproto.PermissionDenied{
		Permission: proto.Uint32(uint32(what)),
		ChannelId:  proto.Uint32(uint32(where.Id)),
		Session:    proto.Uint32(who.Session()),
		Type:       mumbleproto.PermissionDenied_Permission.Enum(),
	}
	err := tlsClient.sendMessage(pd)
	if err != nil {
		//tlsClient.Panicf("%v", err.Error())
		return
	}
}

func (tlsClient *TlsClient) enterChannel(channel *Channel){
	tlsClient.channel = channel
}
