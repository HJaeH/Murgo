
package main
import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/golang/protobuf/proto"
	"../pkg/protobuf"
	"net"
	"io"
	"bufio"
)

type Client struct {
	pid  int64

	username string
	conn net.Conn

	server *Server
	reader  *bufio.Reader

	tcpaddr *net.TCPAddr
	certHash string



	//client auth infomations
	codecs []int32
	tokens []string

	//crypt state
	cryptState CryptState

}

//init client
func (client *Client) init(conn net.Conn){

	client.conn = conn
	client.reader = bufio.NewReader(client.conn)

}

func (client *Client) sendMessage(msg interface{}) error {
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

	_, err = client.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}


/*
// TLS receive loop
func (client *Client) tlsRecvLoop() {
	for {
		// The version handshake is done, the client has been authenticated and it has received
		// all necessary information regarding the server.  Now we're ready to roll!
		// Try to read the next message in the pool
		msg, _ := client.readProtoMessage()

		client.server.incoming <- msg

	}
}*/


func (client *Client) readProtoMessage() (msg *Message, err error) {
	var (
		length uint32
		kind   uint16
	)

	// Read the message type (16-bit big-endian unsigned integer)
	err = binary.Read(client.reader, binary.BigEndian, &kind)
	if err != nil {
		return
	}

	// Read the message length (32-bit big-endian unsigned integer)
	err = binary.Read(client.reader, binary.BigEndian, &length)
	if err != nil {
		return
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(client.reader, buf)
	if err != nil {
		return
	}

	msg = &Message{
		buf:    buf,
		kind:   kind,
		client: client,
	}

	return msg, err
}

//TODO
func (client *Client) Disconnect() {


}

func (client *Client)readPacket(temp []byte, ) (n int, err error){


	n, err = client.conn.Read(temp)

	return n, err
}
