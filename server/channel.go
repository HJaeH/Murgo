package server

import (
	"murgo/pkg/mumbleproto"
	"murgo/pkg/servermodule/log"

	"reflect"

	"time"

	"github.com/golang/protobuf/proto"
)

const ROOT_CHANNEL uint32 = 0

type Channel struct {
	Id       uint32
	Name     string
	children map[int]*Channel
	clients  map[uint32]*Client

	//not used
	Position    int
	temporary   bool
	parentId    uint32
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
	for _, eachUser := range c.clients {
		if reflect.DeepEqual(eachUser, client) {
			continue
		}
		err := client.sendMessage(eachUser.toUserState())
		if err != nil {
			log.Error("Error sending channel User list")
			return err
		}
	}
	return nil
}

func (c *Channel) BroadCastChannel(msg interface{}) {
	//todo : 브로드캐스팅 시 지연 없으면 문제 발생.
	time.Sleep(100 * time.Millisecond)
	for _, eachClient := range c.clients {
		if err := eachClient.sendMessage(msg); err != nil {
			log.Error(err)
		}
	}
}

func (c *Channel) BroadCastChannelWithoutMe(msg interface{}, withoutMe *Client /*, exceptFor ...*Client*/) {
	for _, eachClient := range c.clients {
		if reflect.DeepEqual(withoutMe, eachClient) {
			continue
		}
		if err := eachClient.sendMessage(msg); err != nil {
			log.Error(err)
		}
	}
}

func (c *Channel) broadcastVoice(msg interface{}, withoutMe *Client) {
	for _, eachClient := range c.clients {
		if reflect.DeepEqual(withoutMe, eachClient) {
			continue
		}
		if eachClient.selfDeaf == true {
			continue
		}

		if err := eachClient.sendMessage(msg); err != nil {
			log.Error(msg)
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

func (c *Channel) userCount() int {
	return len(c.clients)
}
