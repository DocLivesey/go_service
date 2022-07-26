package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
)

var build = "develop"

func main() {

	if _, err := maxprocs.Set(); err != nil {
		log.Println("maxprocs [%w]", err)
		os.Exit(1)
	}

	g := runtime.GOMAXPROCS(0)
	log.Printf("starting service build[%s] CPUs[%d]", build, g)
	defer log.Println("service stoped")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Println("service stopppings")
}
