package server

import (
	"murgo/pkg/mumbleproto"

	"github.com/golang/protobuf/proto"

)


//todo : Each Channel should be a module not just a data structure
type Channel struct {
	Id       int
	Name     string
	Position int

	temporary bool
	clients   map[uint32]*TlsClient
	parentId  int
	children  map[int]*Channel
	description string

	// TODO : to be figured out its role
	//rootChannel *Channel
	// Links
	//Links map[int]*Channel
}



func NewChannel(id int, name string) (channel *Channel) {
	channel = new(Channel)
	channel.Id = id
	channel.Name = name
	channel.clients = make(map[uint32]*TlsClient)
	channel.parentId = ROOT_CHANNEL
	channel.Position = 0
	channel.temporary = true
	return channel
}

// TODO : need to be run as genserver
func (channel *Channel)startChannel(){
}

func (channel *Channel) IsEmpty() bool {
	return (len(channel.clients) == 0)
}

func (channel *Channel) removeClient(client *TlsClient) {
	delete(channel.clients, client.Session())
	client.channel = nil
}
func (channel *Channel) addClient(client *TlsClient){
	channel.clients[client.Session()] = client
}

func (channel *Channel) ToChannelState() (*mumbleproto.ChannelState) {
	 channelStateMsg := &mumbleproto.ChannelState{
		ChannelId: proto.Uint32(uint32(channel.Id)),
		Parent: proto.Uint32(uint32(channel.parentId)),
		Name: proto.String(channel.Name),
		Description: proto.String(channel.description),
		Temporary: proto.Bool(channel.temporary),
		Position: proto.Int32(int32(channel.Position)),
	}
	return channelStateMsg
}


