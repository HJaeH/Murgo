package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"murgo/config"
	"murgo/pkg/mumbleproto"

	"github.com/golang/protobuf/proto"
)

type Client struct {
	// 유저가 접속중인 channel
	Channel *Channel
	conn    *net.Conn
	session uint32

	UserName string
	userId   int
	reader   *bufio.Reader

	tcpaddr  *net.TCPAddr
	certHash string

	bandWidth *BandWidth
	//user's setting
	selfDeaf   bool
	selfMute   bool
	mute       bool
	deaf       bool
	tcpPingAvg float32
	tcpPingVar float32
	tcpPackets uint32
	opus       bool
	suppress   bool

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
//func NewTlsClient(conn *net.Conn, session uint32, supervisor *MurgoSupervisor) (*TlsClient){
func NewTlsClient(conn *net.Conn, session uint32) *Client {
	//create new object
	tlsClient := new(Client)
	tlsClient.cryptState = new(config.CryptState)

	//tlsClient.MurgoSupervisor = supervisor
	tlsClient.bandWidth = NewBandWidth()
	tlsClient.conn = conn
	tlsClient.session = session
	tlsClient.reader = bufio.NewReader(*tlsClient.conn)

	tlsClient.testCounter = 0

	// 기본으로 루트채널에 할당
	tlsClient.Channel = nil
	return tlsClient
}

//send msg to client
func (c *Client) SendMessage(msg interface{}) error {

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

	_, err = (*c.conn).Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

//
func (c *Client) readProtoMessage() (msg *Message, err error) {
	var (
		length uint32
		kind   uint16
	)

	// Read the message type (16-bit big-endian unsigned integer)
	//read data form io.reader
	err = binary.Read(c.reader, binary.BigEndian, &kind)
	if err != nil {
		return
	}

	// Read the message length (32-bit big-endian unsigned integer)
	err = binary.Read(c.reader, binary.BigEndian, &length)
	if err != nil {
		return
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(c.reader, buf)
	if err != nil {
		return
	}
	c.testCounter++

	msg = &Message{
		buf:         buf,
		kind:        kind,
		client:      c,
		testCounter: c.testCounter,
	}
	return msg, err
}

func (c *Client) Disconnect() {

}
func (c *Client) ToUserState() *mumbleproto.UserState {

	userStateMsg := &mumbleproto.UserState{
		Session:   proto.Uint32(c.session),
		Name:      proto.String(c.UserName),
		UserId:    proto.Uint32(uint32(c.userId)),
		ChannelId: proto.Uint32(uint32(c.Channel.Id)),
		Mute:      proto.Bool(c.mute),
		Deaf:      proto.Bool(c.deaf),
		Suppress:  proto.Bool(c.suppress),
		SelfDeaf:  proto.Bool(c.selfDeaf),
		SelfMute:  proto.Bool(c.selfMute),
	}
	return userStateMsg
}

func (c *Client) Session() uint32 {
	return c.session
}
