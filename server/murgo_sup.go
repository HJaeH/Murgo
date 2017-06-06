package server

import (
	"errors"
	"fmt"
	"murgo/pkg/servermodule"
	"reflect"
)

type MurgoSupervisor struct {
}

func (murgoSupervisor *MurgoSupervisor) startTlsClient(chan int) {

}

func Start() {
	servermodule.StartSupervisor(new(MurgoSupervisor)) // root

}

func (supervisor *MurgoSupervisor) Terminate() error {
	// todo : terminate
	return errors.New("Error")
}

func Terminate() error {
	err := errors.New("terminate")
	return err
}

func (m *MurgoSupervisor) Temp(a int) {
	fmt.Println(a)
}

//callbacks
func (ms *MurgoSupervisor) Init() {

	servermodule.StartLinkGenServer(ms, new(SessionManager))
	servermodule.StartLinkGenServer(ms, new(ChannelManager))
	servermodule.StartLinkGenServer(ms, new(TlsServer))
	servermodule.StartLinkGenServer(ms, new(MessageHandler))

	inputs := make([]reflect.Value, 1)
	temp := reflect.ValueOf(1123999)
	inputs[0] = temp
	reflect.ValueOf(new(MurgoSupervisor)).MethodByName("Temp").Call(inputs)

	fmt.Println("murgo sup init finished")
}

//todo : startchild 구현, tlsclient -> supervisor로 구현
//todo : startChild -> interface callback
/*
func (ms *MurgoSupervisor) StartChild() {
	servermodule.Call("TlsClient", )
}
*/

/*func (murgoSupervisor *MurgoSupervisor)StartLink(){
	//NewSupervisor calls the init callback
	supervisor := servermodule.StartLinkSupervisor(murgoSupervisor)

	//supervisor.Start(murgoSupervisor)
	murgoSupervisor.supervisor = supervisor
}*/
