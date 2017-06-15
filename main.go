package main

import (
	"fmt"
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
		panic("Error while terminating server")
	}

	fmt.Println("Murgo server terminated")
	os.Exit(1)

}
