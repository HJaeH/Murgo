// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server main
//
//
//


package main

import "murgo/server"

func main() {

	supervisor := server.NewSupervisor()
	go supervisor.StartSupervisor()
	select {}

}
