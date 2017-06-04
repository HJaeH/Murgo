package util

import (
	"net"
)

type Supervisor interface {
	Init()
}

type ChannelManager interface {
	AddChannel(string, uint32)
	BroadCastChannel(int, interface{})
	broadCastChannelWithoutMe(int, msg interface{}, client interface{})
	userEnterChannel() (int, interface{})
	removeChannel(channel interface{})
}

type SessionManager interface {
	broadcastMessage(interface{})
	handleIncomingClient(*net.Conn)
	//getClientList() map[uint32]*server.TlsClient
	getClientBySession(uint32)
}

type MessageHandler interface {
	handleMassage(interface{})
	/*handleAuthenticateMessage(*server.Message)
	handlePingMessage(*server.Message)
	handleChannelRemoveMessage(*server.Message)
	handleChannelStateMessage(*server.Message)
	handleUserStateMessage(*server.Message)
	handleUserRemoveMessage(*server.Message)
	handleBanListMessage(*server.Message)
	handleTextMessage(*server.Message)
	handleAclMessage(*server.Message)
	handleQueryUsers(*server.Message)
	handleCryptSetup(*server.Message)
	handleUserList(*server.Message)
	handleVoiceTarget(*server.Message)
	handlePermissionQuery(*server.Message)
	handleUserStatsMessage(*server.Message)
	handleRequestBlob(*server.Message)*/
}
type TlsServer interface {
	Accept()
}
