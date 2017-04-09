// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server의 main

package main
import (
	_"fmt"
)

const(
	ROOT_SERVER = 0
)

func main() {
	server, err := CreateServer(ROOT_SERVER)
	if err != nil {
		//log.p
	}
	go server.StartServer()
	// keep main goroutine alive
	select {}

}
