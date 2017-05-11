// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server supervisor
// 고루틴을 실행시키고 고루틴간의 통신시 슈퍼바이저를 통해서 각자의 채널에 접근할수 있게 한다


package servermodule

import (
	"fmt"

)

type Supervisor struct {


}


func (supervisor *Supervisor)StartSupervisor() {


}




func (supervisor *Supervisor)StartGenServer(genServer *GenServer) {
	fmt.Println("gen server started")

	// go routine doesn't return here.
	// gen server는 start()에서 리턴 후 deferred recover를 수행해야 하기 때문에 한단계 위에서 goroutine분기 필요 - > 슈퍼바이저에서 처리
	go genServer.start()
}
