package server

import (
	"github.com/golang/protobuf/proto"
	"mumble.info/grumble/pkg/mumbleproto"

)

type Channel struct {
	Id       int
	Name     string
	Position int

	temporary bool
	clients   map[uint32]*TlsClient
	parentId  int
	children  map[int]*Channel
	description string


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
	return channel
}


func (channel *Channel)startChannel(){

}


/*func (channel *Channel) AddChild(child *Channel) {
	child.parent = channel
	channel.children[child.Id] = child
}*/

func (channel *Channel) IsEmpty() bool {
	return (len(channel.clients) == 0)
}

func (channel *Channel) removeClient(client *TlsClient) {
	delete(channel.clients, client.Session())
	client.channel = nil
}
func (channel *Channel) addClient(client *TlsClient){
	channel.clients[client.Session()] = client
	client.channel = channel
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


