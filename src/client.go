
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
	"fmt"
	//"database/sql/driver"
)


type Client struct {
	pid  int64

	username string
	conn net.Conn
	server *Server
	reader  *bufio.Reader
	tcpaddr *net.TCPAddr
	certHash string

	cryptState CryptState

	//state string
}

//init client
func (client *Client) init(conn net.Conn){

	client.conn = conn
	client.reader = bufio.NewReader(client.conn)

}

func (client *Client) sendMessage(msg interface{}, conn net.Conn) error {
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

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}



// TLS receive loop
func (client *Client) tlsRecvLoop() {
	for {
		// The version handshake is done, the client has been authenticated and it has received
		// all necessary information regarding the server.  Now we're ready to roll!
		// Try to read the next message in the pool
		msg, _ := client.readProtoMessage()

		client.server.incoming <- msg

	}
}


func (client *Client) readProtoMessage() (msg *Message, err error) {
	var (
		length uint32
		kind   uint16
	)

	// Read the message type (16-bit big-endian unsigned integer)
	fmt.Println("check 01010")
	err = binary.Read(client.reader, binary.BigEndian, &kind)
	if err != nil {
		return
	}
	fmt.Println("check 01")

	// Read the message length (32-bit big-endian unsigned integer)
	err = binary.Read(client.reader, binary.BigEndian, &length)
	if err != nil {
		return
	}
	fmt.Println("check 02")

	buf := make([]byte, length)
	_, err = io.ReadFull(client.reader, buf)
	if err != nil {
		return
	}
	fmt.Println("check 03")

	msg = &Message{
		buf:    buf,
		kind:   kind,
		client: client,
	}

	return
}

//TODO
func (client *Client) Disconnect() {


}

func (client *Client)readPacket(temp []byte, ) (n int, err error){


	n, err = client.conn.Read(temp)

	return n, err
}
