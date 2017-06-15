package server

import (
	"fmt"

	"murgo/pkg/mumbleproto"

	"reflect"

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

func (c *Channel) removeClient(client *Client) {
	delete(c.clients, client.Session())
	client.Channel = nil
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

func (c *Channel) SendUserListInChannel(client *Client) {
	fmt.Println(c.Name)
	for _, eachUser := range c.clients {
		fmt.Print(eachUser.UserName)
		if reflect.DeepEqual(eachUser, client) {
			continue
		}
		err := client.sendMessage(eachUser.toUserState())
		if err != nil {
			panic(" Error sending channel User list")
		}
	}
}

func (c *Channel) BroadCastChannel(msg interface{}) {
	for _, client := range c.clients {
		client.sendMessageWithInterval(msg)
	}
}
func (c *Channel) BroadCastChannelWithoutMe(me *Client, msg interface{}) {
	for _, eachClient := range c.clients {
		if reflect.DeepEqual(me, eachClient) {
			continue
		}
		eachClient.sendMessageWithInterval(msg)
	}
}

//callback
func (c *Channel) Init() {

}
