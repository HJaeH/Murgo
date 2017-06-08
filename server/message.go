package server

type Message struct {
	buf         []byte
	kind        uint16
	client      *Client
	testCounter int
}

/*
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
*/
