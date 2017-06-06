package main

import (
	"murgo/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	//start supervisor
	server.Start()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // main routine wait for C^ interrupt signal

	err := server.Terminate()
	if err != nil {
		//todo : 비정상 종료
	}
	os.Exit(1)
}
