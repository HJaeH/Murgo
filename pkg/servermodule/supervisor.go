// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor
// 고루틴을 실행시키고 고루틴간의 통신시 슈퍼바이저를 통해서 각자의 채널에 접근할수 있게 한다
package servermodule
import (
	"fmt"

)
//cast, call 통신을 위해서는 모든 genServer에 접근 할 수 있는 모듈이 필요하고
//실시간으로 각 genServer로 부터 cast, call 을 처리 할 수 있어야 함
type Supervisor struct {
	genServer *GenServer
	children map[interface{}] *GenServer
}

func (supervisor *Supervisor)StartSupervisor() {
	//Supervisor itself is a genServer
	supervisor.children = make(map[interface{}] *GenServer)
	supervisor.genServer = new(GenServer)
	supervisor.genServer.init()
	supervisor.genServer.isSupervisor = true
	supervisor.StartGenServer(supervisor.genServer)
}

func (supervisor *Supervisor)StartGenServer(genServer *GenServer) {
	fmt.Println("genServer started")
	//set parent
	genServer.parent = supervisor
	//Register child genServer if it's not a supervisor
	if !genServer.isSupervisor{
		supervisor.children[genServer.resource] = genServer
	}
	go genServer.start()
}

func (supervisor *Supervisor)Terminate(){
	for _, eachChild := range supervisor.children {
		eachChild.callback.Terminate() // todo : terminate를 interface callback 으로?. 서로 같은 동작이면 receiver함수로 하고, 다르면 interface로 구현맞나?
	}
	supervisor.genServer.callback.Terminate()
}

////////////////////////
// internal functions //
////////////////////////

func (supervisor *Supervisor)sendCast(target *GenServer, msg *CastMessage) {
	supervisor.children[target].castChannel <- msg
}
func (supervisor *Supervisor)sendCall(target *GenServer, msg *CallMessage) {
	supervisor.children[target].callChannel <- msg
}
