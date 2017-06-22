package server

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"

	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/server/util/event"
	"murgo/server/util/log"
	"time"
)

const ROOT_CHANNEL uint32 = 0

type ChannelManager struct {
	channelList   map[uint32]*Channel
	nextChannelID uint32
	rootChannel   *Channel
}

func (c *ChannelManager) Init() {
	servermodule.RegisterAPI((*ChannelManager).SendChannelList, event.SendChannelList)
	servermodule.RegisterAPI((*ChannelManager).EnterChannel, event.EnterChannel)
	servermodule.RegisterAPI((*ChannelManager).BroadCastChannel, event.BroadcastChannel)
	servermodule.RegisterAPI((*ChannelManager).AddChannel, event.AddChannel)
	servermodule.RegisterAPI((*ChannelManager).BroadCastChannelWithoutMe, event.BroadCastChannelWithoutMe)
	servermodule.RegisterAPI((*ChannelManager).BroadCastVoiceToChannel, event.BroadCastVoiceToChannel)
	servermodule.RegisterAPI((*ChannelManager).RemoveChannel, event.RemoveChannel)
	c.init()

}

func (c *ChannelManager) init() {
	c.channelList = make(map[uint32]*Channel)

	// set root channel as default channel for all user
	rootChannel := NewChannel(ROOT_CHANNEL, "RootChannel")
	c.rootChannel = rootChannel
	c.channelList[ROOT_CHANNEL] = rootChannel

	//Add one for each channel ID
	c.nextChannelID = ROOT_CHANNEL + 1
}

func (c *ChannelManager) AddChannel(channelName string, client *Client) error {

	if check := c.checkChannelNameDuplication(channelName); check {
		if err := client.sendPermissionDenied(mumbleproto.PermissionDenied_ChannelName); err != nil {
			return log.Error(err)
		}
		return log.Error("Channel name duplicated")
	}

	// create new channel
	channel := NewChannel(c.nextChannelID, channelName)
	c.channelList[channel.Id] = channel

	//let all session know the created channel
	servermodule.Call(event.BroadcastMessage, channel.toChannelState())

	//let the channel creator enter the channel
	c.EnterChannel(channel.Id, client)
	c.nextChannelID++
	return nil

}

func (c *ChannelManager) RootChannel() *Channel {
	return c.channelList[ROOT_CHANNEL]
}

func (c *ChannelManager) BroadCastChannel(channelId uint32, msg interface{}) {

	channel, err := c.channel(channelId)
	if err != nil {
		fmt.Println(err)
	}
	channel.BroadCastChannel(msg)
}

func (c *ChannelManager) BroadCastChannelWithoutMe(channelId uint32, me *Client, msg interface{}) {
	channel, err := c.channel(channelId)
	if err != nil {
		fmt.Println(err)
	}
	channel.BroadCastChannelWithoutMe(msg, me)
}

func (c *ChannelManager) channel(channelId uint32) (*Channel, error) {
	if channel, ok := c.channelList[channelId]; ok {
		return channel, nil
	}

	return nil, errors.New("Channel ID in invalid in channel list")
}

func (c *ChannelManager) SendChannelList1(client *Client) {

	for _, eachChannel := range c.channelList {
		client.sendMessage(eachChannel.toChannelState())
	}
}
func (c *ChannelManager) SendChannelList(client *Client) {
	client.sendMessage(c.RootChannel().toChannelState())
	time.Sleep(100 * time.Millisecond)

	for _, eachChannel := range c.channelList {
		if eachChannel.Id != ROOT_CHANNEL {
			client.sendMessage(eachChannel.toChannelState())
		}

	}
}

//todo : enter channel 이랑 client.disconnect 에서 중복 코드 존재
func (c *ChannelManager) EnterChannel(channelId uint32, client *Client) error {

	newChannel, _ := c.channel(channelId)
	oldChannel := client.Channel
	if oldChannel != nil {
		oldChannel.leave(client)
		if oldChannel.IsEmpty() {
			err := c.RemoveChannel(oldChannel)
			if err != nil {
				return err
			}
		}
	}

	client.Channel = newChannel
	newChannel.addClient(client)

	userState := client.toUserState()

	//새 채널에 알림
	if newChannel.Id != ROOT_CHANNEL {
		//새 채널입장을 채널 유저들에게 알림
		//userState.Mute = proto.Bool(!*userState.ExistUsableMic)
		c.BroadCastChannel(newChannel.Id, userState)
		//채널에 있는 유저들을 입장하는 유저에게 알림
		newChannel.SendUserListInChannel(client)
	} else {
		userState.Mute = proto.Bool(true)
		client.sendMessage(userState)
	}

	return nil

}

func (c *ChannelManager) RemoveChannel(channel *Channel) error {
	// Can't remove root
	if channel.Id == ROOT_CHANNEL {
		return nil
	}

	// move all clients in the channel to root
	for _, client := range channel.clients {
		userStateMsg := &mumbleproto.UserState{}
		userStateMsg.Session = proto.Uint32(client.Session())
		userStateMsg.ChannelId = proto.Uint32(ROOT_CHANNEL)
		c.EnterChannel(ROOT_CHANNEL, client)
		servermodule.AsyncCall(event.BroadcastMessage, userStateMsg)
	}

	// Remove the channel itself
	rootChannel, err := c.channel(ROOT_CHANNEL)
	if err != nil {
		return log.Error("Root doesn't exist")
	}
	delete(c.channelList, channel.Id)
	delete(rootChannel.children, int(channel.Id))
	channelRemoveMsg := &mumbleproto.ChannelRemove{
		ChannelId: proto.Uint32(channel.Id),
	}
	servermodule.AsyncCall(event.BroadcastMessage, channelRemoveMsg)
	return nil
}

func (c *ChannelManager) BroadCastVoiceToChannel(client *Client, voiceData []byte) {
	channel := client.Channel

	c.BroadCastChannelWithoutMe(channel.Id, client, voiceData)
}

func (c *ChannelManager) checkChannelNameDuplication(channelName string) bool {
	for _, eachChannel := range c.channelList {
		if eachChannel.Name == channelName {
			return true
		}
	}
	return false
}
