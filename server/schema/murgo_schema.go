package schema

/*
import (
	"net"
	"bufio"
	"murgo/config"
	"time"
)

type Channel struct {
	Id       int
	Name     string
	Position int

	temporary bool
	clients   map[uint32]*TlsClient
	parentId  int
	children  map[int]*Channel
	description string

	// TODO : not used yet
	//rootChannel *Channel
	// Links
	//Links map[int]*Channel
}


type TlsClient struct {
	// 유저가 접속중인 channel
	Channel     *Channel
	conn        *net.Conn
	session     uint32

	UserName    string
	userId      int
	reader      *bufio.Reader

	tcpaddr     *net.TCPAddr
	certHash    string

	bandWidth   *BandWidth
	//user's setting
	selfDeaf    bool
	selfMute    bool
	mute        bool
	deaf        bool
	tcpPingAvg  float32
	tcpPingVar  float32
	tcpPackets  uint32
	opus        bool
	suppress    bool

	//client auth infomations
	codecs      []int32
	tokens      []string

	//crypt state
	cryptState  *config.CryptState

	//client connection state
	state       int

	//for test
	testCounter int
}


type Message struct {
	buf    []byte
	kind   uint16
	client *TlsClient
	testCounter int
}

type MurgoMessage struct {
	FuncName    string


	Kind        uint16
	ChannelId   int
	Client      *TlsClient
	Channel     *Channel
	Msg         interface{}
	ChannelName string
	Conn        *net.Conn
	CastReply   chan interface{}
}




type BandWidth struct {
	frame_no int
	size_sum int
	size_sum2 int
	bandwidth int

	bandwidth_timer time.Time
	bandwidth_timer2 time.Time

	idle_timer time.Time
	online_timer time.Time
}
*/
