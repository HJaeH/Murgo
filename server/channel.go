package server

import (
	"fmt"

	"murgo/pkg/mumbleproto"
	"murgo/server/util/log"

	"reflect"

	"time"

	"github.com/golang/protobuf/proto"
)

type Channel struct {
	Id       uint32
	Name     string
	Position int

	temporary   bool
	clients     map[uint32]*Client
	parentId    uint32
	children    map[int]*Channel
	description string
}

func NewChannel(id uint32, name string) (channel *Channel) {
	channel = new(Channel)
	channel.Id = id
	channel.Name = name
	channel.clients = make(map[uint32]*Client)
	channel.parentId = ROOT_CHANNEL
	channel.Position = 0
	channel.temporary = true
	return channel
}

func (c *Channel) startChannel() {
}

func (c *Channel) IsEmpty() bool {
	return (len(c.clients) == 0)
}

func (c *Channel) leave(client *Client) {
	userstate := client.toUserState()
	delete(c.clients, client.Session())
	client.Channel = nil
	if c != nil && c.Id != ROOT_CHANNEL {
		c.BroadCastChannel(userstate)
	}

}
func (c *Channel) addClient(client *Client) {
	c.clients[client.Session()] = client
}

func (c *Channel) toChannelState() *mumbleproto.ChannelState {
	channelStateMsg := &mumbleproto.ChannelState{
		ChannelId:   proto.Uint32(c.Id),
		Parent:      proto.Uint32(c.parentId),
		Name:        proto.String(c.Name),
		Description: proto.String(c.description),
		Temporary:   proto.Bool(c.temporary),
		Position:    proto.Int32(int32(c.Position)),
	}
	return channelStateMsg
}

func (c *Channel) SendUserListInChannel(client *Client) error {
	fmt.Println(c.Name)
	for _, eachUser := range c.clients {
		fmt.Print(eachUser.UserName)
		if reflect.DeepEqual(eachUser, client) {
			continue
		}
		err := client.sendMessage(eachUser.toUserState())
		if err != nil {

			client.Disconnect()
			return log.Error("Error sending channel User list")
		}
	}
	return nil
}

func (c *Channel) BroadCastChannel(msg interface{}) {
	//todo : 브로드캐스팅 시 지연 없으면 문제 발생.
	time.Sleep(100 * time.Millisecond)
	for _, client := range c.clients {
		client.sendMessage(msg)
	}
}

func (c *Channel) BroadCastChannelWithoutMe(msg interface{}, withoutMe *Client /*, exceptFor ...*Client*/) {
	for _, eachClient := range c.clients {
		if reflect.DeepEqual(withoutMe, eachClient) {
			continue
		}
		if err := eachClient.sendMessage(msg); err != nil {
			log.ErrorP(msg)
		}
	}
}

func (c *Channel) currentSpeakerCount() int {
	count := 0
	for _, session := range c.clients {
		if session.mute == false {
			count++
		}
	}
	return count
}

//callback
func (c *Channel) Init() {

}
