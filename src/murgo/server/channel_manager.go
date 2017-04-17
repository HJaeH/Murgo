package server





type ChannelManager struct {
	supervisor *Supervisor
	cast chan interface{}


}
type Channel struct {
	Id       int
	parent    *Channel
	Name     string
	Links map[int]*Channel
	Description string
	Position int
}

func (channelManager *ChannelManager) init () {

}
func (channelManager *ChannelManager)startChannelManager(supervisor *Supervisor) {

}