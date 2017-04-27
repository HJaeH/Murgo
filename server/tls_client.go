// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo tls client module

package server


import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"io"
	"bufio"

	"murgo/config"
	"murgo/pkg/mumbleproto"

	"github.com/golang/protobuf/proto"
)


type TlsClient struct {
	supervisor *Supervisor

	// 유저가 접속중인 channel
	channel *Channel
	conn *net.Conn
	session uint32

	userName string
	userId int
	reader  *bufio.Reader

	tcpaddr *net.TCPAddr
	certHash string

	bandWidth *BandWidth
	//user's setting
	selfDeaf bool
	selfMute bool
	mute bool
	deaf bool
	tcpPingAvg float32
	tcpPingVar float32
	tcpPackets uint32
	opus         bool
	suppress bool

	//client auth infomations
	codecs []int32
	tokens []string

	//crypt state
	cryptState *config.CryptState

	//client connection state
	state int

	//for test
	testCounter int


}


// called at session manager
func NewTlsClient(conn *net.Conn, session uint32, supervisor *Supervisor) (*TlsClient){

	//create new object
	tlsClient := new(TlsClient)
	tlsClient.cryptState = new(config.CryptState)

	tlsClient.supervisor = supervisor
	tlsClient.bandWidth = NewBandWidth()
	tlsClient.conn = conn
	tlsClient.session = session
	tlsClient.reader = bufio.NewReader(*tlsClient.conn)

	tlsClient.testCounter = 0

	// 기본으로 루트채널에 할당
	tlsClient.channel = nil
	return tlsClient
}

func (tlsClient *TlsClient) recvLoop(){
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

	_, err = (*tlsClient.conn).Write(buf.Bytes())
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


func (tlsClient *TlsClient) Disconnect() {

}
func (tlsClient *TlsClient) ToUserState()(*mumbleproto.UserState) {

	userStateMsg := &mumbleproto.UserState{
		Session: proto.Uint32(tlsClient.Session()),
		Name: proto.String(tlsClient.userName),
		UserId: proto.Uint32(uint32(tlsClient.userId)),
		ChannelId:proto.Uint32(uint32(tlsClient.channel.Id)),
		Mute:proto.Bool(tlsClient.mute),
		Deaf:proto.Bool(tlsClient.deaf),
		Suppress:proto.Bool(tlsClient.suppress),
		SelfDeaf:proto.Bool(tlsClient.selfDeaf),
		SelfMute: proto.Bool(tlsClient.selfMute),
	}
	return userStateMsg
}


func (tlsClient *TlsClient) Session()(uint32) {
	return tlsClient.session
}
