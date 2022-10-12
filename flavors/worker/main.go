package main

import (
	"os"
	"os/signal"

	"cloud.google.com/go/compute/metadata"
	"github.com/byvko-dev/am-cloud-functions/messaging"
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

	cancel := make(chan int, 1)
	go func() {
		err := stats.StartUpdateWorkers(client, env.MustGetString("PUBSUB_SUBSCRIPTION_ID"), cancel)
		if err != nil {
			panic(err)
		}
	}()

	// Wait for system signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel <- 1
	logs.Info("Shutting down...")
}
