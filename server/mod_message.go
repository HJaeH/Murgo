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
	//todo : refer the usage of murmur idle time
	msg.client.resetIdle()
	var err error
	switch msg.kind {
	case mumbleproto.MessagePing:
		err = messageHandler.handlePingMessage(msg)
	case mumbleproto.MessageUserStats:
		messageHandler.handleUserStatsMessage(msg)

	case mumbleproto.MessageChannelState:
		//msg.client.resetIdle()
		messageHandler.handleChannelStateMessage(msg)
	case mumbleproto.MessageUserState:
		//msg.client.resetIdle()
		messageHandler.handleUserStateMessage(msg)
	case mumbleproto.MessageTextMessage:
		//msg.client.resetIdle()
		messageHandler.handleTextMessage(msg)
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
	client.setPing(ping)
	client.sendMessage(ping)

	return nil

}

func (m *MessageHandler) handleUserStateMessage(msg *Message) error {
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
	// case A. send message to itself 자기자신에게 보내는 경우 - 채널 이동, 상태변경
	if actor.Session() == targetSession {
		target := actor
		//case A1. enter channel
		if userState.ChannelId != nil {
			servermodule.Call(apikeys.EnterChannel, *userState.ChannelId, actor)
			return nil
		}
		//case A2. update my userState in root channel
		if target.Channel.Id == ROOT_CHANNEL {
			if userState.ExistUsableMic != nil &&
				userState.ExistUsableSpeaker != nil {
				//change device state
				target.existUsableMic = userState.GetExistUsableMic()
				target.existUsableSpeaker = userState.GetExistUsableSpeaker()
				if err := target.sendMessage(userState); err != nil {
					log.Error(err)
				}
			} else {
				if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
					log.Error(err)
				}
			}
			return nil
		}
		//case A3. update my userState in normal channel
		if userState.Mute != nil {
			//self mute와 별개로 channel에는 정해진 수의 발언권을 가진 유저들이 있다
			if userState.GetMute() == false {
				if !actor.existUsableMic ||
					!actor.existUsableSpeaker {
					if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
						log.Error(err)
					}
					return nil
				}
				if actor.Channel.currentSpeakerCount() >= config.MaxSpeaker {
					if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
						log.Error(err)
					}
					return nil
				}
				actor.mute = false
				servermodule.AsyncCall(apikeys.BroadcastChannel, actor.Channel.Id, userState)
			} else { // resign the right to speak by itself
				target.mute = true
				servermodule.AsyncCall(apikeys.BroadcastChannel, actor.Channel.Id, userState)

			}
			return nil
		}

		//case A4.change userState by itself.
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
		if changed {
			servermodule.AsyncCall(apikeys.BroadcastChannel, actor.Channel.Id, userState)
		} else {
			if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
				return err
			}
		}

	} else { //case B. send userState to other person (target) -
		/*if userState.GetMute() == true {
			if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
				return err
			}
		}

		// give actors right to talk to target
		if userState.GetMute() == false {
			//only when actor has the right
			if target.user.existUsableMic &&
				target.user.existUsableSpeaker &&
				actor.user.mute == false {
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
				// give it to target
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
	client.sendMessage(msg.client.toUserStats())
}

func (m *MessageHandler) handleTextMessage(msg *Message) {
	client := msg.client
	textMsg := &mumbleproto.TextMessage{}
	err := proto.Unmarshal(msg.buf, textMsg)
	if err != nil {
		panic(err)
		return
	}
	if len(*textMsg.Message) == 0 {
		return
	}
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

func (m *MessageHandler) Init() {
	servermodule.RegisterAPI((*MessageHandler).HandleMessage, apikeys.HandleMessage)

}
