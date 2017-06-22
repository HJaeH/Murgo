package event

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
	RemoveClient
	SendMultipleMessages
	GiveSpeakAbility
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
	RemoveChannel
)

const ( // 401 ~
	Accept = ModServer + (iota + 1)
	Receive
)
