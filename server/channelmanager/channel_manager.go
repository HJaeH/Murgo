// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo channel manager

package channelmanager

import (
	"errors"
	"fmt"
	"reflect"

	"murgo/pkg/mumbleproto"

	"github.com/golang/protobuf/proto"
	"github.com/go-redis/redis"

)

/*type ChannelManager struct {

	supervisor *MurgoSupervisor

	channelList   map[int]*Channel
	nextChannelID int

	Cast chan interface{}
	//Call chan interface{}

	rootChannel *Channel
}*/

const ROOT_CHANNEL = 0

func NewChannelManager(supervisor *MurgoSupervisor) *ChannelManager {

	//assign heap
	channelManager := new(ChannelManager)
	channelManager.channelList = make(map[int]*Channel)
	channelManager.Cast = make(chan interface{})
	//channelManager.Call = make(chan interface{})

	// set uservisor
	channelManager.supervisor = supervisor

	// set root channel as default channel for all user
	rootChannel := NewChannel(ROOT_CHANNEL, "RootChannel")
	channelManager.rootChannel = rootChannel
	channelManager.channelList[ROOT_CHANNEL] = rootChannel

	//Add one for each channel ID
	channelManager.nextChannelID = ROOT_CHANNEL + 1

	return channelManager
}

const ( // TODO : keep other module from accessing those, enum or name space,,,,
	addChannel uint16 = iota
	broadCastChannel
	sendChannelList
	userEnterChannel
)

// channel receiving loop
func (channelManager *ChannelManager) StartChannelManager() {

	// todo : running 추가 또는 젠서버로 통일
	defer func(){
		if err:= recover(); err!= nil{
			fmt.Println("Channel manager recovered")
			channelManager.StartChannelManager()
		}
	}()

	for {
		select {
		case castData := <-channelManager.Cast:
			channelManager.handleCast(castData)
		}
	}
	//channelManager.GenServer.Cast(channelManager.GenServer.)
}

/*
type messenger interface {
	startMessageHandler()
	handleCast1()

}


func (channelManager *ChannelManager) handleCast1 (){
	messenger.handleCast1(3)
}*/
func (channelManager *ChannelManager) handleCast(castData interface{}) {
	murgoMsg := castData.(*MurgoMessage)


	switch murgoMsg.Kind {
	default:
		panic("Handling cast of unexpected type in channel manager")
	case addChannel:
		fmt.Println(murgoMsg.FuncName)
		channelManager.addChannel(murgoMsg.ChannelName, murgoMsg.Client)
	case userEnterChannel:
		channelManager.userEnterChannel(murgoMsg.ChannelId, murgoMsg.Client)
	case broadCastChannel:
		channelManager.broadCastChannel(murgoMsg.ChannelId, murgoMsg.Msg)
	case sendChannelList:
		channelManager.sendChannelList(murgoMsg.Client)

	}
}

func (channelManager *ChannelManager) addChannel(channelName string, client *TlsClient) {
	for _, eachChannel := range channelManager.channelList {
		if eachChannel.Name == channelName {
			sendPermissionDenied(client, mumbleproto.PermissionDenied_ChannelName)
			fmt.Println("duplicated channel name")
			return
		}
	}
	// create new channel
	channel := NewChannel(channelManager.nextChannelID, channelName)
	channelManager.nextChannelID++
	channelManager.channelList[channel.Id] = channel


	//let all session know the created channel
	channelStateMsg := channel.ToChannelState()
	channelManager.supervisor.sessionManager.Cast <- &MurgoMessage{
		Kind: broadcastMessage,
		Msg:  channelStateMsg,
	}

	//let the channel creator enter the channel
	channelManager.userEnterChannel(channel.Id, client)

}

func (channelManager *ChannelManager) RootChannel() *Channel {
	return channelManager.channelList[ROOT_CHANNEL]
}

func (channelManager *ChannelManager) exitChannel(client *TlsClient, channel *Channel) {

}

//broadcast a msg to all users in a channel
func (channelManager *ChannelManager) broadCastChannel(channelId int, msg interface{}) {
	channel, err := channelManager.channel(channelId)
	if err != nil {
		fmt.Println(err)
	}
	for _, client := range channel.clients {
		client.SendMessage(msg)
	}
}

func (channelManager *ChannelManager) broadCastChannelWithoutMe(channelId int, msg interface{}, client *TlsClient) {
	channel, err := channelManager.channel(channelId)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(channel.clients)
	for _, eachClient := range channel.clients {
		if reflect.DeepEqual(client, eachClient) {
			continue
		}
		eachClient.SendMessage(msg)
	}
}

func (channelManager *ChannelManager) channel(channelId int) (*Channel, error) {
	if channel, ok := channelManager.channelList[channelId]; ok {
		return channel, nil
	}

	return nil, errors.New("Channel ID in invalid in channel list")
}

func (channelManager *ChannelManager) sendChannelList(client *TlsClient) {
	fmt.Println(len(channelManager.channelList))
	for _, eachChannel := range channelManager.channelList {

		client.SendMessage(eachChannel.ToChannelState())
	}
}

func (channelManager *ChannelManager) userEnterChannel(channelId int, client *TlsClient) {

	newChannel, err := channelManager.channel(channelId)
	fmt.Println(client.UserName, " will enter ", newChannel.Name)
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
			channelManager.removeChannel(oldChannel)
			oldChannel = nil
		}
	}

	client.Channel = newChannel
	newChannel.addClient(client)
	userState := client.ToUserState()


	if oldChannel != nil && oldChannel.Id != ROOT_CHANNEL {
		//이전 채널에 떠났음을 알림
		channelManager.broadCastChannelWithoutMe(oldChannel.Id, userState, client)
	}
	// 변한 상태를 클라이언트에게 알림


	if newChannel.Id != ROOT_CHANNEL {
		//새 채널입장을 채널 유저들에게 알림
		channelManager.broadCastChannelWithoutMe(newChannel.Id, userState, client)
		//채널에 있는 유저들을 입장하는 유저에게 알림
		newChannel.sendUserListInChannel(client)
	/*	for _, users := range newChannel.clients {
			client.sendMessage(users.ToUserState())

		}*/
		//client.SendMessage(userState)
	} else {
		fmt.Println("-------------------")
		client.SendMessage(userState)
	}

	//for test
	for _, eachChannel := range channelManager.channelList{
		fmt.Print(eachChannel.Name, ": ")
		for _, eachUser := range eachChannel.clients {
			fmt.Print(eachUser.UserName, ", ")
		}
		fmt.Println()
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

		//channelManager.Call(channelManager.supervisor.sessionManager)

		channelManager.supervisor.sessionManager.Cast <- &MurgoMessage{
			Kind: broadcastMessage,
			Msg:  userStateMsg,
		}
	}

	// Remove the channel itself
	rootChannel, err := channelManager.channel(ROOT_CHANNEL)
	if err != nil {
		panic("Root doesn't exist")
	}
	delete(channelManager.channelList, channel.Id)
	delete(rootChannel.children, channel.Id)

	channelRemoveMsg := &mumbleproto.ChannelRemove{
		ChannelId: proto.Uint32(uint32(channel.Id)),
	}
	channelManager.supervisor.sessionManager.Cast <- &MurgoMessage{
		Kind: broadcastMessage,
		Msg:  channelRemoveMsg,
	}
}

func (channelManager *ChannelManager) printChannels() {
	fmt.Println("channel list : ")
	for _, channel := range channelManager.channelList {
		fmt.Print(channel.Name, ", ")
	}
	fmt.Println()
}

/*

func (genServer *GenServer) handleCast(castData interface{}){
	temp := castData.(CastMessage)
	funcName := temp.funcName
	args := temp.args

	inputs := make([]reflect.Value, len(args))

	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])

	}
	reflect.ValueOf(genServer).MethodByName(funcName).Call(inputs)
}

func (genServer *GenServer) handleCall(msg *CallMessage){
	funcName := msg.funcName
	args := msg.args
	returnChan := msg.returnChan

	inputs := make([]reflect.Value, len(args))

	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])

	}
	reflect.ValueOf(genServer).MethodByName(funcName).Call(inputs)
	returnChan <- &CallReply{
		//sender: ,

	} // todo : call return 작업중
}
*/


func ExampleClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
	err = client.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exists")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exists
}
