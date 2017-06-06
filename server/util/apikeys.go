package util

const (
	SessionManager = (iota + 1) * 100
	MessageHandler
	ChannelManager
	TlsServer

	//
)

const (
	//101 ~
	HandleIncomingClient = SessionManager + iota + 1
)

const (
	//201 ~
	HandleMessage = MessageHandler + iota + 1
)

const (
	//301 ~
	SendChannellist = ChannelManager + iota + 1
	UserEnterChannel
	BroadcastMessage
	BroadcastChannel
	AddChannel
)

const ( // 401 ~
	Accept = TlsServer + (iota + 1)
)
