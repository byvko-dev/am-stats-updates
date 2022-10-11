package main

import (
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/byvko-dev/am-cloud-functions/messaging"
	"github.com/byvko-dev/am-cloud-functions/scheduler"
	"github.com/byvko-dev/am-cloud-functions/stats"
	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
)

func main() {
	// Try to get project ID from env first and then from metadata
	projectID := os.Getenv("GCLOUD_PROJECT")
	if projectID == "" {
		id, err := metadata.ProjectID()
		if err != nil {
			panic(err)
		}
		projectID = id
	}
	topicID := env.MustGetString("PUBSUB_TOPIC_ID")

	// Messaging client
	client, err := messaging.NewClient(projectID, topicID)
	if err != nil {
		panic(err)
	}

	flavor := env.MustGetString("FLAVOR")
	if flavor == "scheduler" {
		runScheduler(client)
	}
	if flavor == "worker" {
		runWorker(client)
	}
	logs.Info("All tasks completed, exiting")
	os.Exit(0)
}

func runScheduler(client *messaging.Client) {
	realm := env.MustGetString("TASK_REALM")
	// Publish tasks messages
	err := scheduler.CreateRealmTasks(client, scheduler.TaskTypeSnapshot, realm, 3) // 3 tries per task
	if err != nil {
		panic(err)
	}
	err = scheduler.CreateRealmTasks(client, scheduler.TaskTypeAccountUpdate, realm, 3) // 3 tries per task
	if err != nil {
		panic(err)
	}

	logs.Info("Done creating tasks for realm: %s", realm)
}

func runWorker(client *messaging.Client) {
	timeoutStr := env.MustGetString("SUBSCRIPTION_TIMEOUT_MIN")
	timeoutMin, err := strconv.Atoi(timeoutStr)
	if err != nil {
		panic(err)
	}
	timeout := time.Minute * time.Duration(timeoutMin)
	stats.StartUpdateWorkers(client, env.MustGetString("PUBSUB_SUBSCRIPTION_ID"), timeout)
}
