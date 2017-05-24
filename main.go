// @author 허재화 <jhwaheo@smilegate.com>
// @version 1.0
// murgo server main
//
//
//


package main

import (
	"murgo/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	murgoSupervisor := new(server.MurgoSupervisor)

	//start supervisor
	murgoSupervisor.Start()

	server.StartSupervisor()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // main routine wait for C^ interrupt signal

	err := murgoSupervisor.Terminate()
	if err != nil {
		//todo : 비정상 종료
	}
	os.Exit(1)
}
