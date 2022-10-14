package main

import (
	"os"
	"os/signal"

	"github.com/byvko-dev/am-cloud-functions/scheduler"
	"github.com/byvko-dev/am-cloud-functions/stats"
	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
	"github.com/robfig/cron/v3"
)

func main() {
	cancel := make(chan int, 1)
	go func() {
		err := stats.StartUpdateWorkers(cancel) // Will execute all tasks in the queue every 5 min
		if err != nil {
			panic(err)
		}
	}()

	withCron := env.MustGetString("WITH_CRON")
	if withCron == "true" {
		runner := cron.New()
		// Update players and sessions
		runner.AddFunc("0 9 * * *", func() { createRealmTasks("NA") })    // NA
		runner.AddFunc("0 1 * * *", func() { createRealmTasks("EU") })    // EU
		runner.AddFunc("0 18 * * *", func() { createRealmTasks("ASIA") }) // ASIA
		runner.Start()
		defer runner.Stop()
	}

	// Wait for system signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel <- 1
	logs.Info("Shutting down...")
}

func createRealmTasks(realm string) {
	err := scheduler.CreateRealmTasks(scheduler.TaskTypeSnapshot, realm, 3) // 3 tries per task
	if err != nil {
		logs.Error("Error creating snapshot tasks for realm: %s", realm)
	}
	err = scheduler.CreateRealmTasks(scheduler.TaskTypeAccountUpdate, realm, 3) // 3 tries per task
	if err != nil {
		logs.Error("Error creating account update tasks for realm: %s", realm)
	}
}
