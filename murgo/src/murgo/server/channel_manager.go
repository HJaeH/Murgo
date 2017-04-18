package server

import (
	"murgo/data"
	"fmt"
)

type ChannelManager struct {
	supervisor *Supervisor
	cast chan interface{}

	channelList map[int]*data.Channel
	nextChannelID int


}


func NewChannelManager(supervisor *Supervisor)(*ChannelManager){

	channelManager := new(ChannelManager)
	channelManager.nextChannelID = 0
	channelManager.cast = make(chan interface{})
	channelManager.supervisor = supervisor
	channelManager.channelList = make(map [int]*data.Channel)


	//channelid 생성마다 증가
	channelManager.nextChannelID++

	return channelManager
}



func (channelManager *ChannelManager)startChannelManager(supervisor *Supervisor) {

}

func (channelManager *ChannelManager)handleCast( castData interface{}) {
	switch t := castData.(type) {
	default:
		fmt.Printf("unexpected type %T", t)
	case string:
		//msg := castData.(string)
		//todo
	}


}
/*

func (channelManager *ChannelManager) AddChannel(name string) (channel *data.Channel) {
	channel = data.NewChannel(channelManager.nextChanId, name)
	server.Channels[channel.Id] = channel
	server.nextChanId += 1

	return
}*/
