package scheduler

import (
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-core/logs"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddQueueItems(item ...helpers.Payload) error {
	return database.AddQueueItems(item...)
}

func UpdateQueueItem(item helpers.Payload) error {
	return database.UpdateQueueItem(item)
}

func MarkComplete(item helpers.Payload) error {
	item.IsProcessing = false
	item.TriesLeft = 0
	return UpdateQueueItem(item)
}

func ExecuteAllQueueTasks(concurrency int, expiration time.Duration, handler func(helpers.Payload)) {
	limiter := make(chan int, concurrency)
	for {
		limiter <- 1
		go func() {
			data, err := database.GetNextTask(expiration)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					logs.Error("Failed to get next task: %s", err.Error())
				}
				time.Sleep(1 * time.Minute)
			} else {
				logs.Info("Executing queue task %s", data.ID.Hex())
				handler(data)
			}
			<-limiter
		}()
	}
}
