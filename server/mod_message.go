package server

import (
	"fmt"
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/server/util/apikeys"
	"murgo/server/util/log"

	"murgo/config"

	"github.com/golang/protobuf/proto"
)

type MessageHandler struct{}

type Message struct {
	buf    []byte
	kind   uint16
	client *Client
}

func (messageHandler *MessageHandler) HandleMessage(msg *Message) {
	var err error
	switch msg.kind {
	case mumbleproto.MessagePing:
		err = messageHandler.handlePingMessage(msg)
	case mumbleproto.MessageChannelRemove:
		messageHandler.handleChannelRemoveMessage(msg)
	case mumbleproto.MessageChannelState:
		messageHandler.handleChannelStateMessage(msg)
	case mumbleproto.MessageUserState:
		messageHandler.handleUserStateMessage(msg)
	case mumbleproto.MessageUserRemove:
		messageHandler.handleUserRemoveMessage(msg)
	case mumbleproto.MessageTextMessage:
		messageHandler.handleTextMessage(msg)
	case mumbleproto.MessageUserStats:
		messageHandler.handleUserStatsMessage(msg)

	/* 	todo : 코드 정리
	case mumbleproto.MessageACL:
		messageHandler.handleAclMessage(msg)
	case mumbleproto.MessageQueryUsers:
		messageHandler.handleQueryUsers(msg)
	case mumbleproto.MessageCryptSetup:
		messageHandler.handleCryptSetup(msg)
	//case mumbleproto.MessageContextAction:
	//	fmt.Print("MessageContextAction from client")
	case mumbleproto.MessageUserList:
		messageHandler.handleUserList(msg)
	case mumbleproto.MessageVoiceTarget:
		messageHandler.handleVoiceTarget(msg)
	case mumbleproto.MessagePermissionQuery:
		messageHandler.handlePermissionQuery(msg)
	case mumbleproto.MessageRequestBlob:
		messageHandler.handleRequestBlob(msg)
	case mumbleproto.MessageBanList:
		messageHandler.handleBanListMessage(msg)
	*/
	default:
		err = log.Error("Uncategorized msg type", msg.kind)
	}
	if err != nil {
		log.ErrorP(err)
	}
}

func (m *MessageHandler) handlePingMessage(msg *Message) error {
	ping := &mumbleproto.Ping{}
	err := proto.Unmarshal(msg.buf, ping)
	if err != nil {
		return err
	}
	client := msg.client
	client.sendMessage(&mumbleproto.Ping{
		Timestamp:  ping.Timestamp,
		TcpPackets: ping.TcpPackets,
		TcpPingVar: ping.TcpPingVar,
		TcpPingAvg: ping.TcpPingAvg,
	})

	return nil

}

func (m *MessageHandler) handleUserStateMessage(msg *Message) error {

	// 메시지를 보낸 유저 reset idle -> 이 부분은 통합
	userState := &mumbleproto.UserState{}
	err := proto.Unmarshal(msg.buf, userState)
	if err != nil {
		return err
	}
	fmt.Println(userState)

	actor := msg.client
	targetSession := userState.GetSession()

	// deaf, suppress, priority_speaker는 변경할 수 없다.
	if userState.Deaf != nil ||
		userState.Suppress != nil ||
		userState.PrioritySpeaker != nil {
		if err := actor.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
			return err
		}
		return nil
	}
	// case A. 자기자신에게 보내는 경우 - 채널 이동, 상태변경
	if actor.Session() == targetSession {
		target := actor
		//case A1. enter channel
		if userState.ChannelId != nil {
			//todo: call or asyncall
			servermodule.AsyncCall(apikeys.EnterChannel, *userState.ChannelId, actor)
			return nil
		}
		//case A2. update my userState in root channel
		if target.Channel.Id == ROOT_CHANNEL {
			if userState.ExistUsableMic != nil &&
				userState.ExistUsableSpeaker != nil {
				// 디바이스 상태를 갱신할 수 있다.
				// 응답은 액터에게만 한다.
				target.existUsableMic = userState.GetExistUsableMic()
				target.existUsableSpeaker = userState.GetExistUsableSpeaker()

				// 유저에게 변경된 유저상태 전송
				if err := target.sendMessageWithInterval(userState); err != nil {
					log.Error(err)
				}
			} else {
				// 디바이스 상태 이외의 갱신은 허가하지 않는다.
				if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
					log.Error(err)
				}
			}
			return nil
		}
		//case A3. update my userState in normal channel
		if userState.Mute != nil {
			// 나의 말하기 권한을 획득하려고 할 경우,
			if userState.GetMute() == false {
				// 나의 마이크와 스피커가 사용가능해위야 하며,
				if !actor.existUsableMic ||
					!actor.existUsableSpeaker {
					if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
						log.Error(err)
					}
					return nil
				}
				// 채널의 말하기 권한 유저수가 최대치를 넘지 않았을 경우 가능하다.
				if actor.Channel.currentSpeakerCount() >= config.MaxSpeaker {
					if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
						log.Error(err)
					}
					return nil
				}
				// 말하기 권한 획득
				actor.mute = false
				servermodule.AsyncCall(apikeys.BroadcastChannel, actor.Channel.Id, userState)
			} else {
				// 나의 말하기 권한 포기
				target.mute = true
				servermodule.AsyncCall(apikeys.BroadcastChannel, actor.Channel.Id, userState)

			}
			return nil
		}

		//case A4.
		changed := false
		if userState.ExistUsableMic != nil {
			target.existUsableMic = userState.GetExistUsableMic()
			changed = true
		}
		if userState.ExistUsableSpeaker != nil {
			target.existUsableSpeaker = userState.GetExistUsableSpeaker()
			changed = true
		}
		if userState.SelfDeaf != nil {
			target.selfDeaf = userState.GetSelfDeaf()
			changed = true
		}
		if userState.SelfMute != nil {
			target.selfMute = userState.GetSelfMute()
			changed = true
		}

		// 채널에 있는 모든 유저에게 변경된 상태 브로드캐스트
		if changed {
			servermodule.AsyncCall(apikeys.BroadcastChannel, actor.Channel.Id, userState)
		} else {
			if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
				return err
			}
		}

	} else { //case B. 타인에게 보내는 경우 -
		/*if userState.GetMute() == true {
			if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
				return err
			}
		}

		// 다른 유저에게 말하기 권한을 양도할 수 있다.
		if userState.GetMute() == false {
			// 내가 말하기 권한이 있는 경우에만 양도 가능
			if target.user.existUsableMic &&
				target.user.existUsableSpeaker &&
				actor.user.mute == false {
				// 나의 말하기 권한을 막고
				actor.user.mute = true
				newUserState := &mumble.UserState{
					Session: proto.Uint32(actor.sessionID),
					Actor:   proto.Uint32(actor.sessionID),
					Mute:    proto.Bool(true),
				}
				if err := o.server.channelManager.BroadcastToChannel(actor.channel, newUserState, nil); err != nil {
					log.Error(err)
					return
				}
				// 다른 유저의 말하기 권한 할당
				target.user.mute = true
				newUserState.Session = proto.Uint32(target.sessionID)
				newUserState.Actor = proto.Uint32(target.sessionID)
				newUserState.Mute = proto.Bool(false)
				if err := o.server.channelManager.BroadcastToChannel(target.channel, newUserState, nil); err != nil {
					log.Error(err)
					return
				}
			}
			return
		}*/
	}

	return nil

}

//구현된 핸들링 함수
func (m *MessageHandler) handleChannelStateMessage(tempMsg interface{}) {
	msg := tempMsg.(*Message)
	channelStateMsg := &mumbleproto.ChannelState{}
	err := proto.Unmarshal(msg.buf, channelStateMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ChannelState info:", channelStateMsg, "from:", msg.client.UserName)
	if channelStateMsg.ChannelId == nil && channelStateMsg.Name != nil && *channelStateMsg.Temporary == true && *channelStateMsg.Parent == 0 && *channelStateMsg.Position == 0 {
		servermodule.Call(apikeys.AddChannel, *channelStateMsg.Name, msg.client)
	}
}
func (m *MessageHandler) handleUserStatsMessage(msg *Message) {
	userStats := &mumbleproto.UserStats{}
	client := msg.client
	err := proto.Unmarshal(msg.buf, userStats)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("userstats info:", userStats, "from:", msg.client.UserName)
	newUserStatsMsg := &mumbleproto.UserStats{
		TcpPingAvg: proto.Float32(client.tcpPingAvg),
		TcpPingVar: proto.Float32(client.tcpPingVar),
		Opus:       proto.Bool(client.opus),
		//TODO : elapsed time 계산 과정 추가해서 idle, online 시간 추적
		//Bandwidth:
		//Onlinesecs:
		//Idlesecs:
	}
	client.sendMessage(newUserStatsMsg)
}

// protocol handling dummy for UserRemoveMessage
func (messageHandler *MessageHandler) handleUserRemoveMessage(msg *Message) {
	msgProto := &mumbleproto.UserRemove{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("userRemoveMessage info:", msgProto, "from:", msg.client.UserName)
}

//TODO : deal with sending text message
func (m *MessageHandler) handleTextMessage(msg *Message) {
	client := msg.client
	textMsg := &mumbleproto.TextMessage{}
	err := proto.Unmarshal(msg.buf, textMsg)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("Text msg info:", textMsg, "from:", msg.client.UserName)

	//todo : 예외처리
	/*if len(textMsg.Message) == 0 {
		return
	}*/
	newMsg := &mumbleproto.TextMessage{
		Actor:   proto.Uint32(client.Session()),
		Message: textMsg.Message,
	}
	// send text message to channels
	for _, eachChannelId := range textMsg.ChannelId {
		servermodule.AsyncCall(apikeys.BroadCastChannelWithoutMe, eachChannelId, client, newMsg)
	}

	// send text message to users
	servermodule.AsyncCall(apikeys.SendMessages, textMsg.Session, newMsg)
}

// protocol handling dummy for Aclmessage

func (m *MessageHandler) handleAclMessage(msg *Message) {
	msgProto := &mumbleproto.ACL{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ACL info:", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for QueryUSers

func (m *MessageHandler) handleQueryUsers(msg *Message) {
	msgProto := &mumbleproto.QueryUsers{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Handle query users  :", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for banlist message
func (m *MessageHandler) handleBanListMessage(msg *Message) {

	msgProto := &mumbleproto.BanList{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Banlist message info:", msgProto, "from:", msg.client.UserName)
}

// protocol handling dummy for CtyptSetup
func (m *MessageHandler) handleCryptSetup(msg *Message) {
	msgProto := &mumbleproto.CryptSetup{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("handle cryptsetup :", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for UserList

func (m *MessageHandler) handleUserList(msg *Message) {
	msgProto := &mumbleproto.UserList{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(" userlist :", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for Voicetarget
func (m *MessageHandler) handleVoiceTarget(msg *Message) {
	msgProto := &mumbleproto.VoiceTarget{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("voice target info:", msgProto, "from:", msg.client.UserName)
}

// protocol handling dummy for PermissionQuery
func (m *MessageHandler) handlePermissionQuery(msg *Message) {
	msgProto := &mumbleproto.PermissionQuery{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("permission query info:", msgProto, "from:", msg.client.UserName)

}

// protocol handling dummy for requestBlob
func (m *MessageHandler) handleRequestBlob(msg *Message) {
	msgProto := &mumbleproto.RequestBlob{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("requestblob info:", msgProto, "from:", msg.client.UserName)

}

func (m *MessageHandler) handleChannelRemoveMessage(msg *Message) {
	msgProto := &mumbleproto.RequestBlob{}
	err := proto.Unmarshal(msg.buf, msgProto)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("channel remove info:", msgProto, "from:", msg.client.UserName)
}

func (m *MessageHandler) Init() {
	servermodule.RegisterAPI((*MessageHandler).HandleMessage, apikeys.HandleMessage)

}
