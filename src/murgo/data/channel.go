package data

import (
	"murgo/server"
)

type Channel struct {
	Id       int
	Name     string
	Position int

	temporary bool
	clients   map[uint32]*server.TlsClient
	parent    *Channel
	children  map[int]*Channel

	// Links
	Links map[int]*Channel

}

func NewChannel(id int, name string) (channel *Channel) {
	channel = new(Channel)
	channel.Id = id
	channel.Name = name
	channel.clients = make(map[uint32]*server.TlsClient)
	channel.children = make(map[int]*Channel)
	channel.Links = make(map[int]*Channel)
	return
}

func (channel *Channel) AddChild(child *Channel) {
	child.parent = channel
	channel.children[child.Id] = child
}

func (channel *Channel) RemoveChild(child *Channel) {
	child.parent = nil
	delete(channel.children, child.Id)
}


// Returns a slice of all channels in this channel's link
func (channel *Channel) AllLinks() (seen map[int]*Channel) {
	seen = make(map[int]*Channel)
	walk := []*Channel{channel}
	for len(walk) > 0 {
		current := walk[len(walk)-1]
		walk = walk[0 : len(walk)-1]
		for _, linked := range current.Links {
			if _, alreadySeen := seen[linked.Id]; !alreadySeen {
				seen[linked.Id] = linked
				walk = append(walk, linked)
			}
		}
	}
	return
}


func (channel *Channel) IsTemporary() bool {
	return channel.temporary
}

func (channel *Channel) IsEmpty() bool {
	return len(channel.clients) == 0
}
/*

func (channel *Channel) AddClient(client *server.TlsClient) {
	channel.clients[client.Session()] = client
	client.Channel = channel
}

func (channel *Channel) RemoveClient(client *Client) {
	delete(channel.clients, client.Session())
	client.Channel = nil
}
*/


//