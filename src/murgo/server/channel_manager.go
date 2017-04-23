package server

import (
	"fmt"
	"errors"
)

type ChannelManager struct {
	supervisor *Supervisor


	channelList map[int]*Channel
	nextChannelID int

	Cast chan interface{}
	Call chan interface{}

	rootChannel *Channel
}




const ROOT_CHANNEL = 0

func NewChannelManager(supervisor *Supervisor)(*ChannelManager){

	channelManager := new(ChannelManager)

	channelManager.rootChannel = new(Channel)
	channelManager.rootChannel.Id = ROOT_CHANNEL


	channelManager.Cast = make(chan interface{})
	channelManager.supervisor = supervisor


	channelManager.channelList = make(map [int]*Channel)
	//0번 채널은 루트
	channelManager.channelList[0] = NewChannel(0, "Root")
	// 다음채널은 1부터 시작
	channelManager.nextChannelID = 1

	return channelManager
}



// channel receiving loop
func (channelManager *ChannelManager)startChannelManager(supervisor *Supervisor) {
	for{
		select {
		case castData := <-channelManager.Cast:
			channelManager.handleCast(castData)
		}
	}
}



const (
	addChannel uint16 = iota
	enterChannel
	broadCastChannel
)



func (channelManager *ChannelManager)handleCast( castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch  murgoMsg.kind {
	default:
		fmt.Printf("unexpected type ")
	case addChannel:
		channelManager.addChannel(murgoMsg.ChannelName)
	case enterChannel:
		channelManager.enterChannel(murgoMsg.client, murgoMsg.channel)
	}
}



// APIs
func (channelManager *ChannelManager) addChannel(name string) (channel *Channel) {
	channel = NewChannel(channelManager.nextChannelID, name)
	channelManager.channelList[channel.Id] = channel
	//채널 아이디는 1씩 증가
	channelManager.nextChannelID += 1

	return
}


func (channelManager *ChannelManager) RootChannel()(*Channel) {
	return channelManager.channelList[0]
}


func (channelManager *ChannelManager) enterChannel(client *TlsClient, channel *Channel) {
	channel.addClient(client)

}


func (channelManager *ChannelManager) exitChannel(client *TlsClient, channel *Channel) {

}


//broadcast a msg to all users in a channel
func (channelManager *ChannelManager) broadCastChannel(channelId int, msg *Message){
	channel, err := channelManager.channel(channelId);
	if err != nil {
		fmt.Println(err)
	}
	for _, client := range channel.clients { //다른 루틴 데이터 접근 read 작업
		client.sendMessage(msg)
	}
}



////////Internal functions

func (channelManager *ChannelManager)toChannelStateMsg(channel *Channel)(msg *Message){
	return &Message{}
}




// return channel
func (channelManager *ChannelManager)channel(channelId int) (*Channel, error){
	if channel, ok := channelManager.channelList[channelId]; ok {
		return channel, nil
	}

	return nil, errors.New("Channel ID in invalid in channel list")
}




