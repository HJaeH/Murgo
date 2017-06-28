package main

import (
	"murgo/pkg/servermodule/log"
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
		log.Error(err, "Error while terminating server")
		os.Exit(1)
	}

	log.Info("Murgo server terminated")
	log.Info("-------------------------")
	os.Exit(1)
}
