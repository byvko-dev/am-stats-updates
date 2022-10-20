package scheduler

import (
	"encoding/json"

	"github.com/byvko-dev/am-stats-updates/internal/core/database"
	"github.com/byvko-dev/am-stats-updates/internal/core/helpers"
	"github.com/byvko-dev/am-stats-updates/internal/core/messaging"
)

func SaveUpdateResults(results []helpers.UpdateResult, taskType string) error {
	return database.UpsertUpdateLogs(results, taskType)
}

func AddQueueItem(item helpers.UpdateTask) error {
	task, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return messaging.SendQueueMessage(messaging.QueueCacheUpdates, task)
}

func SubscribeToTasks(concurrency int, handler func(helpers.UpdateTask) error, cancel chan int) error {
	handlerWrapper := func(payload []byte) error {
		var task helpers.UpdateTask
		err := json.Unmarshal(payload, &task)
		if err != nil {
			return err
		}
		return handler(task)
	}

	return messaging.SubscribeToQueue(messaging.QueueCacheUpdates, handlerWrapper, concurrency, cancel)
}
