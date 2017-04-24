package server

import (
	"fmt"
	"errors"
	"mumble.info/grumble/pkg/mumbleproto"
	"github.com/golang/protobuf/proto"

	"reflect"
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
	channelManager.channelList = make(map [int]*Channel)
	channelManager.Cast = make(chan interface{})
	channelManager.Call = make(chan interface{})


	channelManager.supervisor = supervisor
	rootChannel := NewChannel(ROOT_CHANNEL, "RootChannel")
	channelManager.rootChannel = rootChannel
	channelManager.channelList[ROOT_CHANNEL] = rootChannel
	// 다음채널은 1부터 시작
	channelManager.nextChannelID = ROOT_CHANNEL + 1

	return channelManager
}

const ( // enum 이나 name space
	addChannel uint16 = iota
	broadCastChannel
	sendChannelList
	userEnterChannel
)

// channel receiving loop
func (channelManager *ChannelManager)startChannelManager() {
	for{
		select {
		case castData := <-channelManager.Cast:
			channelManager.handleCast(castData)
		}
	}
}
func (channelManager *ChannelManager)handleCast( castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)

	switch  murgoMsg.kind {
	default:
		fmt.Printf("unexpected type cm ")
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

// APIs
func (channelManager *ChannelManager) addChannel(channelName string, client *TlsClient) () {
	for _, eachChannel := range channelManager.channelList {
		if eachChannel.Name == channelName {
			sendPermissionDenied(client, mumbleproto.PermissionDenied_ChannelName)
			fmt.Println("duplicated channel name")
			return
		}
	}
	channel := NewChannel(channelManager.nextChannelID, channelName)
	channelManager.channelList[channel.Id] = channel
	//channel.Position = *(int32(channelStateMsg.Position))
	//channel.temporary = *channelStateMsg.Temporary




	channelStateMsg := channel.ToChannelState()
	channelManager.supervisor.sm.Cast <- &MurgoMessage{
		kind:broadcastMessage,
		msg:channelStateMsg,
	}


	channelManager.sendChannelList(client)
	channelManager.broadCastChannel(channel.Id, channelStateMsg)
	channelManager.userEnterChannel(channel.Id, client)

	return
}


func (channelManager *ChannelManager) RootChannel()(*Channel) {
	return channelManager.channelList[ROOT_CHANNEL]
}



func (channelManager *ChannelManager) exitChannel(client *TlsClient, channel *Channel) {

}


//broadcast a msg to all users in a channel
func (channelManager *ChannelManager) broadCastChannel(channelId int, msg interface{}){
	channel, err := channelManager.channel(channelId);
	if err != nil {
		fmt.Println(err)
	}
	for _, client := range channel.clients { //다른 루틴 데이터 접근 read 작업
		client.sendMessage(msg)
	}
}

func (channelManager *ChannelManager) broadCastChannelWithoutMe(channelId int, msg interface{}, client *TlsClient){
	channel, err := channelManager.channel(channelId);
	if err != nil {
		fmt.Println(err)
	}
	for _, eachClient := range channel.clients { //다른 루틴 데이터 접근 read 작업
		if reflect.DeepEqual(client, eachClient) {
			continue
		}
		eachClient.sendMessage(msg)
	}
}

func (channelManager *ChannelManager)channel(channelId int) (*Channel, error){
	if channel, ok := channelManager.channelList[channelId]; ok {
		return channel, nil
	}

	return nil, errors.New("Channel ID in invalid in channel list")
}

func (channelManager *ChannelManager) sendChannelList(client *TlsClient) {
	fmt.Println(len(channelManager.channelList))
	for _, eachChannel := range channelManager.channelList {

		client.sendMessage(eachChannel.ToChannelState())
	}
}


func (channelManager *ChannelManager) userEnterChannel(channelId int, client *TlsClient){

	newChannel, err := channelManager.supervisor.cm.channel(channelId) //todo
	if err != nil {
		panic("Channel Id doesn't exist")
	}
	oldChannel := client.channel
	if oldChannel == newChannel {
		return
	}
	if oldChannel != nil {
		oldChannel.removeClient(client)
		if oldChannel.IsEmpty() {
			channelManager.removeChannel(oldChannel)

		}
	}
	newChannel.addClient(client)

	userState := client.ToUserState()
	if oldChannel != nil && oldChannel.Id != ROOT_CHANNEL {
		//이전 채널에 떠났음을 알림
		channelManager.broadCastChannelWithoutMe(oldChannel.Id, userState, client)
	}
	// 변한 상태를 클라이언트에게 알림
	client.sendMessage(client.ToUserState())


	if newChannel.Id != ROOT_CHANNEL {
		//새 채널입장을 알림
		channelManager.broadCastChannelWithoutMe(newChannel.Id, userState, client)
	}

}




func (channelManager *ChannelManager) removeChannel(channel *Channel) {

	// Can't remove root
	if channel.Id == ROOT_CHANNEL {
		return
	}

	// Remove all clients in the channel to root
	for _, client := range channel.clients {

		userStateMsg := &mumbleproto.UserState{}
		userStateMsg.Session = proto.Uint32(client.Session())
		userStateMsg.ChannelId = proto.Uint32(uint32(ROOT_CHANNEL))
		channelManager.userEnterChannel(ROOT_CHANNEL, client)

		channelManager.supervisor.sm.Cast <- &MurgoMessage{
			kind:broadcastMessage,
			msg:userStateMsg,
		}
	}

	// Remove the channel itself
	parent := channelManager.channelList[channel.parentId]
	delete(channelManager.channelList, channel.Id)
	delete(parent.children, channel.Id)


	channelRemoveMsg := &mumbleproto.ChannelRemove{
		ChannelId: proto.Uint32(uint32(channel.Id)),
	}
	channelManager.supervisor.sm.Cast <- &MurgoMessage{
		kind:broadcastMessage,
		msg:channelRemoveMsg,
	}
}
