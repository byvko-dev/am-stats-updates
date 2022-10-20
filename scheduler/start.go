package scheduler

import (
	"context"
	"strings"

	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
	"github.com/robfig/cron/v3"
)

func Start() func() context.Context {
	nameSlice := strings.Split(env.MustGetString("POD_NAME"), "-")
	withCron := nameSlice[len(nameSlice)-1] == "0"
	if withCron {
		logs.Info("Starting cron jobs")
		runner := cron.New()
		// Update players and sessions
		runner.AddFunc("0 0 * * *", createGlossaryTasks)                  // Glossary update
		runner.AddFunc("0 9 * * *", func() { createRealmTasks("NA") })    // NA
		runner.AddFunc("0 1 * * *", func() { createRealmTasks("EU") })    // EU
		runner.AddFunc("0 18 * * *", func() { createRealmTasks("ASIA") }) // ASIA
		runner.Start()
		return runner.Stop
	}
	return func() context.Context { return context.Background() }
}

func createGlossaryTasks() {
	err := CreateGlossaryTasks(TaskTypeUpdateGlossaryVehicles, 2)
	if err != nil {
		logs.Error("Error creating glossary tasks: %s", err.Error())
	}
	err = CreateGlossaryTasks(TaskTypeUpdateGlossaryAchievements, 2)
	if err != nil {
		logs.Error("Error creating glossary tasks: %s", err.Error())
	}
	err = CreateGlossaryTasks(TaskTypeUpdateGlossaryAverages, 2)
	if err != nil {
		logs.Error("Error creating glossary tasks: %s", err.Error())
	}
}

func createRealmTasks(realm string) {
	err := CreateRealmTasks(TaskTypeAccountUpdate, realm, 3, 100) // 3 tries per task
	if err != nil {
		logs.Error("Error creating account update tasks for realm: %s", realm)
	}
	err = CreateRealmTasks(TaskTypeSnapshot, realm, 3, 50) // 3 tries per task
	if err != nil {
		logs.Error("Error creating snapshot tasks for realm: %s", realm)
	}
}
