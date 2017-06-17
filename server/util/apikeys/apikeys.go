package apikeys

const (
	ModSession = (iota + 1) * 100
	ModMessage
	ModChannel
	ModServer
)

const (
	//101 ~
	HandleIncomingClient = ModSession + iota + 1
	BroadcastMessage
	SetUserOption
	RemoveClient
	SendMessages
)

const (
	//201 ~
	HandleMessage = ModMessage + iota + 1
)

const (
	//301 ~
	SendChannelList = ModChannel + iota + 1
	EnterChannel
	BroadcastChannel
	AddChannel
	BroadCastChannelWithoutMe
	BroadCastVoiceToChannel
)

const ( // 401 ~
	Accept = ModServer + (iota + 1)
	Receive
)
