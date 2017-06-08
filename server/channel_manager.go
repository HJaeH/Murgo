package server

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"

	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule"
	APIkeys "murgo/server/util"
	"reflect"
)

const ROOT_CHANNEL = 0

type ChannelManager struct {

	//todo add numChannel and keep channel ids
	channelList   map[int]*Channel
	nextChannelID int
	rootChannel   *Channel
}

func (c *ChannelManager) Init() {
	servermodule.RegisterAPI((*ChannelManager).SendChannelList, APIkeys.SendChannelList)
	servermodule.RegisterAPI((*ChannelManager).EnterChannel, APIkeys.EnterChannel)
	servermodule.RegisterAPI((*ChannelManager).BroadCastChannel, APIkeys.BroadcastChannel)
	servermodule.RegisterAPI((*ChannelManager).AddChannel, APIkeys.AddChannel)
	//assign heap

	c.init()

}
func (c *ChannelManager) init() {
	c.channelList = make(map[int]*Channel)

	// set root channel as default channel for all user
	rootChannel := NewChannel(ROOT_CHANNEL, "RootChannel")
	c.rootChannel = rootChannel
	c.channelList[ROOT_CHANNEL] = rootChannel

	//Add one for each channel ID
	c.nextChannelID = ROOT_CHANNEL + 1
}

const ( // TODO : keep other module from accessing those, enum or name space,,,,
	addChannel uint16 = iota
	broadCastChannel
	sendChannelList
	userEnterChannel
)

// channel receiving loop
func (channelManager *ChannelManager) startChannelManager() {

	// panic 발생시 모든 모듈의 이 시점으로 리턴할 것
	// TODO : 일단 에러 발생 시점 파악을 위해 주석처리 이후에 슈퍼바이저에서 코드 통합 강구
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Channel manager recovered")
			channelManager.startChannelManager()
		}
	}()

	for {
		select {
		case castData := <-channelManager.Cast:
			channelManager.handleCast(castData)
		}
	}
}

// TODO : cast data could be a function .
func (channelManager *ChannelManager) handleCast(castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch murgoMsg.kind {
	default:
		panic("Handling cast of unexpected type in channel manager")
	case addChannel:
		channelManager.addChannel(murgoMsg.ChannelName, murgoMsg.client)
	case userEnterChannel:
		channelManager.userEnterChannel(murgoMsg.channelId, murgoMsg.client)
	case broadCastChannel:
		channelManager.broadCastChannel(murgoMsg.channelId, murgoMsg.msg)
	case sendChannelList:
		channelManager.sendChannelList(murgoMsg.client)
	}

}

func (channelManager *ChannelManager) addChannel(channelName string, client *TlsClient) {
	for _, eachChannel := range channelManager.channelList {
		if eachChannel.Name == channelName {
			//todo : client object 없에는 과정
			//sendPermissionDenied(client, mumbleproto.PermissionDenied_ChannelName)
			fmt.Println("duplicated channel name")
			return
		}
	}
	// create new channel
	channel := NewChannel(channelManager.nextChannelID, channelName)
	channelManager.nextChannelID++
	channelManager.channelList[channel.Id] = channel

	//let all session know the created channel
	channelStateMsg := channel.toChannelState()

	servermodule.Cast(APIkeys.BroadcastMessage, channelStateMsg)
	//let the channel creator enter the channel
	c.EnterChannel(channel.Id, client)

}

func (c *ChannelManager) RootChannel() *Channel {
	return c.channelList[ROOT_CHANNEL]
}

func (c *ChannelManager) exitChannel(client *Client, channel *Channel) {

}

//broadcast a msg to all users in a channel

func (c *ChannelManager) BroadCastChannel(channelId int, msg interface{}) {
	channel, err := c.channel(channelId)
	if err != nil {
		fmt.Println(err)
	}
	for _, client := range channel.clients {
		client.SendMessage(msg)
	}
}

func (channelManager *ChannelManager) broadCastChannelWithoutMe(channelId int, msg interface{}, client *Client) {
	channel, err := channelManager.channel(channelId)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(channel.clients)
	for _, eachClient := range channel.clients {
		if reflect.DeepEqual(client, eachClient) {
			continue
		}
		eachClient.SendMessage(msg)
	}
}

func (c *ChannelManager) channel(channelId int) (*Channel, error) {
	if channel, ok := c.channelList[channelId]; ok {
		return channel, nil
	}

	return nil, errors.New("Channel ID in invalid in channel list")
}

func (c *ChannelManager) SendChannelList(client *Client) {
	//fmt.Println(len(c.channelList))
	for _, eachChannel := range c.channelList {

		client.SendMessage(eachChannel.toChannelState())
	}
}

func (c *ChannelManager) EnterChannel(channelId int, client *Client) {
	fmt.Println("----------")
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

	client.channel = newChannel
	newChannel.addClient(client)
	userState := client.ToUserState()

	if oldChannel != nil && oldChannel.Id != ROOT_CHANNEL {
		//이전 채널에 떠났음을 알림
		c.broadCastChannelWithoutMe(oldChannel.Id, userState, client)
	}
	// 변한 상태를 클라이언트에게 알림

	if newChannel.Id != ROOT_CHANNEL {
		//새 채널입장을 채널 유저들에게 알림
		c.broadCastChannelWithoutMe(newChannel.Id, userState, client)
		//채널에 있는 유저들을 입장하는 유저에게 알림
		newChannel.SendUserListInChannel(client)
		/*	for _, users := range newChannel.clients {
			client.sendMessage(users.ToUserState())

		}*/
		//client.SendMessage(userState)
	} else {
		client.SendMessage(userState)
	}

	//for test
	for _, eachChannel := range c.channelList {
		fmt.Print(eachChannel.Name, ": ")
		for _, eachUser := range eachChannel.clients {
			fmt.Print(eachUser.UserName, ", ")
		}
		fmt.Println()
	}
}

func (c *ChannelManager) removeChannel(tempChannel interface{}) {
	channel := tempChannel.(*Channel)
	// Can't remove root
	if channel.Id == ROOT_CHANNEL {
		return
	}

	// Remove all clients in the channel to root
	for _, client := range channel.clients {
		userStateMsg := &mumbleproto.UserState{}
		userStateMsg.Session = proto.Uint32(client.Session())
		userStateMsg.ChannelId = proto.Uint32(uint32(ROOT_CHANNEL))
		c.EnterChannel(ROOT_CHANNEL, client)

		//channelManager.Call(channelManager.supervisor.sessionManager)
		servermodule.Cast(APIkeys.BroadcastMessage, userStateMsg)
	}

	// Remove the channel itself
	rootChannel, err := c.channel(ROOT_CHANNEL)
	if err != nil {
		panic("Root doesn't exist")
	}
	delete(c.channelList, channel.Id)
	delete(rootChannel.children, channel.Id)

	channelRemoveMsg := &mumbleproto.ChannelRemove{
		ChannelId: proto.Uint32(uint32(channel.Id)),
	}
	servermodule.Cast(APIkeys.BroadcastMessage, channelRemoveMsg)
}

func (c *ChannelManager) printChannels() {
	fmt.Println("channel list : ")
	for _, channel := range c.channelList {
		fmt.Print(channel.Name, ", ")
	}
	fmt.Println()
}
