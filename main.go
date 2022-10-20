package main

import (
	"os"
	"os/signal"

	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/scheduler"
	"github.com/byvko-dev/am-stats-updates/webserver"
	"github.com/byvko-dev/am-stats-updates/workers"
)

func main() {
	// Task workers
	cancel := make(chan int, 1)
	go workers.Start(cancel)

	// Task scheduler
	stopScheduler := scheduler.Start()

	// Web server
	go webserver.Start()

	// Wait for system signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel <- 1
	stopScheduler()
	logs.Info("Shutting down...")
}
