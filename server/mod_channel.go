package server

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"

	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	"murgo/server/util/apikeys"
	"murgo/server/util/log"

	"reflect"
)

const ROOT_CHANNEL uint32 = 0

type ChannelManager struct {

	//todo add numChannel and keep channel ids
	channelList   map[uint32]*Channel
	nextChannelID uint32
	rootChannel   *Channel
}

func (c *ChannelManager) Init() {
	servermodule.RegisterAPI((*ChannelManager).SendChannelList, apikeys.SendChannelList)
	servermodule.RegisterAPI((*ChannelManager).EnterChannel, apikeys.EnterChannel)
	servermodule.RegisterAPI((*ChannelManager).BroadCastChannel, apikeys.BroadcastChannel)
	servermodule.RegisterAPI((*ChannelManager).AddChannel, apikeys.AddChannel)
	servermodule.RegisterAPI((*ChannelManager).BroadCastChannelWithoutMe, apikeys.BroadCastChannelWithoutMe)
	servermodule.RegisterAPI((*ChannelManager).BroadCastVoiceToChannel, apikeys.BroadCastVoiceToChannel)
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
	}

	// create new channel
	channel := NewChannel(c.nextChannelID, channelName)
	c.channelList[channel.Id] = channel

	//let all session know the created channel
	servermodule.AsyncCall(apikeys.BroadcastMessage, channel.toChannelState())

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
	for _, client := range channel.clients {
		//todo : send msg 1
		client.sendMessageWithInterval(msg)
	}
}

func (c *ChannelManager) BroadCastChannelWithoutMe(channelId uint32, me *Client, msg interface{}) {
	channel, err := c.channel(channelId)
	if err != nil {
		fmt.Println(err)
	}
	for _, eachClient := range channel.clients {
		if reflect.DeepEqual(me, eachClient) {
			continue
		}
		//todo send message1
		err = eachClient.sendMessageWithInterval(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (c *ChannelManager) channel(channelId uint32) (*Channel, error) {
	if channel, ok := c.channelList[channelId]; ok {
		return channel, nil
	}

	return nil, errors.New("Channel ID in invalid in channel list")
}

func (c *ChannelManager) SendChannelList(client *Client) {
	for _, eachChannel := range c.channelList {
		client.sendMessage(eachChannel.toChannelState())
	}
}

func (c *ChannelManager) EnterChannel(channelId uint32, client *Client) {
	newChannel, err := c.channel(channelId)
	//fmt.Println(client.UserName, " will enter ", newChannel.Name)
	if err != nil {
		panic("Channel Id doesn't exist")
	}
	oldChannel := client.Channel

	if oldChannel == newChannel {
		return
	}

	if oldChannel != nil {
		oldChannel.removeClient(client)
		if oldChannel.IsEmpty() {
			c.removeChannel(oldChannel)
			oldChannel = nil
		}
	}

	client.Channel = newChannel
	newChannel.addClient(client)
	userState := client.toUserState()

	if oldChannel != nil && oldChannel.Id != ROOT_CHANNEL {
		//이전에 있던 채널에 떠난 유저의 userstate message 전송
		c.BroadCastChannelWithoutMe(oldChannel.Id, client, userState)
		//c.BroadCastChannel(oldChannel.Id, userState)
	}

	//새 채널에 알림
	if newChannel.Id != ROOT_CHANNEL {
		//새 채널입장을 채널 유저들에게 알림

		c.BroadCastChannel(newChannel.Id, userState)
		//c.broadCastChannelWithoutMe(newChannel.Id, userState, client)
		//채널에 있는 유저들을 입장하는 유저에게 알림
		newChannel.SendUserListInChannel(client)
	} else {
		//send message1 따로 분리
		client.sendMessageWithInterval(userState)
	}

	if err != nil {

		fmt.Println("error sending message")
	}

}

func (c *ChannelManager) removeChannel(tempChannel interface{}) {
	channel := tempChannel.(*Channel)
	// Can't remove root
	if channel.Id == ROOT_CHANNEL {
		return
	}

	// move all clients in the channel to root
	for _, client := range channel.clients {
		userStateMsg := &mumbleproto.UserState{}
		userStateMsg.Session = proto.Uint32(client.Session())
		userStateMsg.ChannelId = proto.Uint32(ROOT_CHANNEL)
		c.EnterChannel(ROOT_CHANNEL, client)
		servermodule.AsyncCall(apikeys.BroadcastMessage, userStateMsg)
	}

	// Remove the channel itself
	rootChannel, err := c.channel(ROOT_CHANNEL)
	if err != nil {
		log.Panic("Root doesn't exist")
	}
	delete(c.channelList, channel.Id)
	delete(rootChannel.children, int(channel.Id))
	channelRemoveMsg := &mumbleproto.ChannelRemove{
		ChannelId: proto.Uint32(channel.Id),
	}
	servermodule.AsyncCall(apikeys.BroadcastMessage, channelRemoveMsg)
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

func (c *ChannelManager) printChannels() {
	fmt.Println("channel list : ")
	for _, channel := range c.channelList {
		fmt.Print(channel.Name, ", ")
	}
	fmt.Println()
}
