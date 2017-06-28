package server

import (
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/pkg/servermodule/log"

	"murgo/config"

	"murgo/server/util/event"

	"errors"

	"github.com/golang/protobuf/proto"
)

const (
	MaxTextMessageLength = 200
)

type MessageHandler struct{}

type Message struct {
	buf    []byte
	kind   uint16
	client *Client
}

func (messageHandler *MessageHandler) HandleMessage(msg *Message) {

	client := msg.client
	client.resetIdle()
	var err error
	switch msg.kind {
	case mumbleproto.MessagePing:
		err = messageHandler.handlePingMessage(msg)
	case mumbleproto.MessageUserStats:
		err = messageHandler.handleUserStatsMessage(msg)
	case mumbleproto.MessageChannelState:
		err = messageHandler.handleChannelStateMessage(msg)
	case mumbleproto.MessageUserState:
		err = messageHandler.handleUserStateMessage(msg)
	case mumbleproto.MessageTextMessage:
		err = messageHandler.handleTextMessage(msg)
	default:
		log.Errorf("Uncategorized msg type : %d", msg.kind)
	}
	if err != nil {
		log.Error(err, "error while handling message :", msg.kind)
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

	actor := msg.client
	targetSession := userState.GetSession()

	//  cannot change follows : deaf, suppress, priority_speaker
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
			servermodule.Call(event.EnterChannel, *userState.ChannelId, actor)
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
		//case A3. change my speak ablility
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
					if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_SpeakerFull); err != nil {
						log.Error(err)
					}
					return nil
				}
				target.mute = false

				servermodule.AsyncCall(event.BroadcastChannel, actor.Channel.Id, userState)
			} else { // resign the right to speak by itself
				actor.mute = true
				servermodule.AsyncCall(event.BroadcastChannel, actor.Channel.Id, userState)

			}
			return nil
		}

		//case A4.change userState itself.

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

			//todo :deaf시에 mute도 같이 변함
			// target.selfMute = userState.GetSelfDeaf()
			changed = true
		}
		if userState.SelfMute != nil {
			target.selfMute = userState.GetSelfMute()
			changed = true
		}
		if changed {
			servermodule.AsyncCall(event.BroadcastChannel, target.Channel.Id, userState)
		} else {
			if err := target.sendPermissionDenied(mumbleproto.PermissionDenied_Permission); err != nil {
				return err
			}
		}

	} else { //case B. send userState to other person (target)
		servermodule.Call(event.GiveSpeakAbility, userState)
	}

	return nil

}

func (m *MessageHandler) handleChannelStateMessage(msg *Message) error {
	channelStateMsg := &mumbleproto.ChannelState{}
	err := proto.Unmarshal(msg.buf, channelStateMsg)
	if err != nil {
		return err
	}
	if channelStateMsg.ChannelId == nil && channelStateMsg.Name != nil && *channelStateMsg.Temporary == true && *channelStateMsg.Parent == 0 && *channelStateMsg.Position == 0 {
		servermodule.Call(event.AddChannel, *channelStateMsg.Name, msg.client)
	}

	return nil
}

func (m *MessageHandler) handleUserStatsMessage(msg *Message) error {
	userStats := &mumbleproto.UserStats{}
	client := msg.client
	err := proto.Unmarshal(msg.buf, userStats)
	if err != nil {
		return err
	}
	client.sendMessage(msg.client.toUserStats())
	return nil
}

func (m *MessageHandler) handleTextMessage(msg *Message) error {
	client := msg.client
	textMsg := &mumbleproto.TextMessage{}
	err := proto.Unmarshal(msg.buf, textMsg)
	if err != nil {
		return err
	}
	if len(*textMsg.Message) == 0 {
		return errors.New("No message passed")
	} else if len(*textMsg.Message) > MaxTextMessageLength {
		client.sendPermissionDenied(mumbleproto.PermissionDenied_TextTooLong)
		return nil
	}

	// send text message to channels
	newMsg := &mumbleproto.TextMessage{
		Actor:   proto.Uint32(client.Session()),
		Message: textMsg.Message,
	}
	for _, eachChannelId := range textMsg.ChannelId {
		servermodule.AsyncCall(event.BroadCastChannelWithoutMe, eachChannelId, client, newMsg)
	}

	// send text message to users
	servermodule.AsyncCall(event.SendMultipleMessages, textMsg.Session, newMsg)

	return nil
}

func (m *MessageHandler) Init() {
	servermodule.EventRegister((*MessageHandler).HandleMessage, event.HandleMessage)

}
