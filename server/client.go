package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"murgo/pkg/mumbleproto"
	"murgo/server/util/apikeys"

	"murgo/pkg/servermodule"

	"fmt"

	"github.com/golang/protobuf/proto"
)

type Client struct {
	session  uint32
	UserName string
	Channel  *Channel

	selfDeaf           bool
	selfMute           bool
	mute               bool
	deaf               bool
	prioritySpeaker    bool
	channelOwner       bool
	existUsableMic     bool
	existUsableSpeaker bool

	tcpPingAvg       float32
	tcpPingVar       float32
	tcpPackets       uint32
	opus             bool
	suppress         bool
	onlinesecs       uint32
	bandwidth        uint32
	bandwidthRecord  *Bandwidth
	bandwidthRecord2 *Bandwidth
	idelsecs         uint32

	conn         *net.Conn
	reader       *bufio.Reader
	crypt        *CryptState
	codecs       []int32
	tokens       []string
	disconnected bool
	version      uint32
}

func newClient(conn *net.Conn, session uint32) *Client {
	client := new(Client)
	client.crypt = new(CryptState)
	client.session = session

	client.bandwidthRecord = newBandwidth()
	client.bandwidthRecord2 = newBandwidth()

	client.conn = conn
	client.session = session
	client.reader = bufio.NewReader(*client.conn)

	// 기본으로 루트채널에 할당
	client.Channel = nil
	return client
}

//send msg to client

func (c *Client) sendMessage(msg interface{}) error {

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
		(*c.conn).Close()
		servermodule.AsyncCall(apikeys.RemoveClient, c)
		fmt.Println("clinet left")
	}
}

func (c *Client) toUserState() *mumbleproto.UserState {

	userStateMsg := &mumbleproto.UserState{
		Session:   proto.Uint32(c.session),
		Name:      proto.String(c.UserName),
		UserId:    proto.Uint32(uint32(c.session)),
		ChannelId: proto.Uint32(c.Channel.Id),
		Mute:      proto.Bool(c.mute),
		Deaf:      proto.Bool(c.deaf),
		Suppress:  proto.Bool(c.suppress),
		SelfDeaf:  proto.Bool(c.selfDeaf),
		SelfMute:  proto.Bool(c.selfMute),
	}
	return userStateMsg
}

func (c *Client) toUserStats() *mumbleproto.UserStats {
	msg := &mumbleproto.UserStats{
		TcpPackets: proto.Uint32(c.tcpPackets),
		TcpPingAvg: proto.Float32(c.tcpPingAvg),
		TcpPingVar: proto.Float32(c.tcpPingVar),
		Bandwidth:  proto.Uint32(c.bandwidth),
		Opus:       proto.Bool(c.opus),
		Onlinesecs: proto.Uint32(c.getOnlinesecs()),
		Idlesecs:   proto.Uint32(c.getIdlesecs()),
	}
	return msg
}

func (c *Client) Session() uint32 {
	return c.session
}

func (c *Client) reject(rejectType mumbleproto.Reject_RejectType, reason string) {
	var reasonString *string = nil
	if len(reason) > 0 {
		reasonString = proto.String(reason)
	}
	c.sendMessage(&mumbleproto.Reject{
		Type:   rejectType.Enum(),
		Reason: reasonString,
	})
	c.Disconnect()
}

func (c *Client) sendPermissionDenied(denyType mumbleproto.PermissionDenied_DenyType) error {
	permissionDeniedMsg := &mumbleproto.PermissionDenied{
		Session: proto.Uint32(c.Session()),
		Type:    &denyType,
	}
	fmt.Println("Permission denied", permissionDeniedMsg)
	return c.sendMessage(permissionDeniedMsg)
}

func (c *Client) setPing(pingMsg *mumbleproto.Ping) {
	c.tcpPackets = pingMsg.GetTcpPackets()
	c.tcpPingAvg = pingMsg.GetTcpPingAvg()
	c.tcpPingVar = pingMsg.GetTcpPingVar()
}

func (c *Client) getIdlesecs() uint32 {
	return uint32((nowMicrosec() - c.bandwidthRecord.idleTimer) / 1000000)
}

func (c *Client) getOnlinesecs() uint32 {
	return uint32((nowMicrosec() - c.bandwidthRecord.onlineTimer) / 1000000)
}

func (c *Client) resetIdle() {
	c.bandwidthRecord.idleTimer = nowMicrosec()
}

func (c *Client) addFrame(packetSize uint32) error {
	now := nowMicrosec()
	elapsed := uint64(now-c.bandwidthRecord.bandwidthTimer) / 1000000.0 // sec
	if elapsed == 0 {
		return nil
	}

	calcBandwidth(c.bandwidthRecord, packetSize)
	c.bandwidth = c.bandwidthRecord.bandwidth

	if c.bandwidthRecord.frameNo == HalfFrameSlots {
		c.bandwidthRecord2.reset()
	}
	if c.bandwidthRecord.frameNo >= HalfFrameSlots {
		calcBandwidth(c.bandwidthRecord2, packetSize)
	}
	if c.bandwidthRecord.frameNo == MaxFrameSlots {
		c.bandwidthRecord.copyFrom(c.bandwidthRecord2)
		c.bandwidthRecord2.reset()
	}

	return nil
}
