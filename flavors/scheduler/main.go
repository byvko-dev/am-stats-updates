package main

import (
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/byvko-dev/am-cloud-functions/messaging"
	"github.com/byvko-dev/am-cloud-functions/scheduler"
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

	// Publish tasks messages
	realm := env.MustGetString("TASK_REALM")
	err = scheduler.CreateRealmTasks(client, scheduler.TaskTypeSnapshot, realm, 3) // 3 tries per task
	if err != nil {
		panic(err)
	}
	err = scheduler.CreateRealmTasks(client, scheduler.TaskTypeAccountUpdate, realm, 3) // 3 tries per task
	if err != nil {
		panic(err)
	}

	logs.Info("Done creating tasks for realm: %s", realm)
}
