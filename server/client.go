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
	APIkeys "murgo/server/util"

	"time"

	"murgo/pkg/servermodule"

	"fmt"

	"github.com/golang/protobuf/proto"
)

type Client struct {
	Channel *Channel
	conn    *net.Conn
	session uint32

	UserName string
	userId   uint32
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
	state        int
	disconnected bool

	//version
	version uint32
	//for test
	testCounter int
}

// called at session manager
//func NewTlsClient(conn *net.Conn, session uint32, supervisor *MurgoSupervisor) (*TlsClient){
func newClient(conn *net.Conn, session uint32) *Client {
	//create new object
	client := new(Client)
	client.cryptState = new(config.CryptState)
	client.userId = session

	//tlsClient.MurgoSupervisor = supervisor
	client.bandWidth = NewBandWidth()
	client.conn = conn
	client.session = session
	client.reader = bufio.NewReader(*client.conn)

	client.testCounter = 0

	// 기본으로 루트채널에 할당
	client.Channel = nil
	return client
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
		fmt.Println("-----")
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
	//mutex.Lock()
	_, err = (*c.conn).Write(buf.Bytes())
	if err != nil {
		return err
	}
	//mutex.Unlock()

	return nil
}

//todo : write시에 패킷 loss발생 해결 위함
func (c *Client) SendMessage1(msg interface{}) error {

	//userState := msg.(*mumbleproto.UserState)
	//fmt.Println("send user msg !! userstate info:", userState, "1!!")
	time.Sleep(100 * time.Millisecond)
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
	msg = &Message{
		buf:    buf,
		kind:   kind,
		client: c,
	}
	return msg, err
}

func (c *Client) Disconnect() {
	if !c.disconnected {
		c.disconnected = true
		servermodule.Call(APIkeys.RemoveClient, c)
		(*c.conn).Close()
	}
}

func (c *Client) toUserState() *mumbleproto.UserState {

	userStateMsg := &mumbleproto.UserState{
		Session:   proto.Uint32(c.session),
		Name:      proto.String(c.UserName),
		UserId:    proto.Uint32(uint32(c.userId)),
		ChannelId: proto.Uint32(c.Channel.Id),
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
func (c *Client) Receive() {
	for {
		msg, err := c.readProtoMessage()
		if err != nil {
			if err != nil {
				if err == io.EOF {
					c.Disconnect()
				} else {
					panic(err)
				}
				return
			}
		}
		servermodule.Cast(APIkeys.HandleMessage, msg)
	}
}
