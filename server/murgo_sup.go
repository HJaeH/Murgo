// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor


package server

import (
	"errors"
	"murgo/pkg/servermodule"
)

type MurgoSupervisor struct {
	//supervisor *servermodule.Supervisor
	moduleName string
}
func (murgoSupervisor *MurgoSupervisor)startTlsClient(chan int){
	//servermodule.
}

func StartSupervisor(){
	servermodule.StartSupervisor(&MurgoSupervisor{})
}



func (supervisor *MurgoSupervisor)Terminate() (error) {
	// todo : terminate
	return errors.New("Error")
}


//callbacks
func (murgoSupervisor *MurgoSupervisor)init() ([]interface{}){
	//todo : module list 반환 코드 리팩토링
	modules := make(map[string] interface{})
	modules["sessionManager"] = sessionmanager.StartLink()
	return modules
}


//todo : startchild 구현, tlsclient -> supervisor로 구현
func StartChild(){
	servermodule.Call("TlsClient", &servermodule.CallMessage{})
}


/*func (murgoSupervisor *MurgoSupervisor)StartLink(){
	//NewSupervisor calls the init callback
	supervisor := servermodule.StartLinkSupervisor(murgoSupervisor)

	//supervisor.Start(murgoSupervisor)
	murgoSupervisor.supervisor = supervisor
}*/