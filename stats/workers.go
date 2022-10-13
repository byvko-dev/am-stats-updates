package stats

import (
	"fmt"
	"strconv"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-cloud-functions/scheduler"
	snapshots "github.com/byvko-dev/am-cloud-functions/stats/save-snapshots"
	accounts "github.com/byvko-dev/am-cloud-functions/stats/update-accounts"
	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
	"github.com/robfig/cron/v3"
)

func StartUpdateWorkers(cancel chan int) error {
	workersNub := env.MustGetString("CONCURRENT_WORKERS")
	concurrency, _ := strconv.Atoi(workersNub)
	if concurrency < 1 {
		concurrency = 1
	}

	executeTasks(concurrency)()

	runner := cron.New()
	runner.AddFunc("*/5 * * * * *", executeTasks(concurrency))

	<-cancel
	runner.Stop()
	return nil
}

func executeTasks(concurrency int) func() {
	handlerWrapper := func(p helpers.Payload) {
		err := handler(p)
		if err != nil {
			logs.Error("failed to process payload during cleanup: %v", err)
		}
	}

	return func() {
		logs.Debug("Executing all current tasks")
		scheduler.ExecuteAllQueueTasks(concurrency, time.Minute*60, handlerWrapper)
	}
}

func handler(payload helpers.Payload) error {
	switch payload.Type {
	case scheduler.TaskTypeAccountUpdate:
		result, retryIds, err := accounts.UpdateSomePlayers(payload.Realm, payload.PlayerIDs)

		insertErr := database.AddUpdateLogs(result, payload.Type)
		if insertErr != nil {
			logs.Error("failed to add update logs: %v", insertErr.Error())
		}

		if len(retryIds) > 0 && payload.TriesLeft > 0 {
			err := sendRetryMessage(payload.Type, payload.Realm, retryIds, payload.TriesLeft-1)
			if err != nil {
				return fmt.Errorf("failed to send retry message: %w", err)
			}
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to update accounts: %w", err)
		}
		return scheduler.MarkComplete(payload)

	case scheduler.TaskTypeSnapshot:
		result, retryIds, err := snapshots.SaveAccountSnapshots(payload.Realm, payload.PlayerIDs, false)

		insertErr := database.AddUpdateLogs(result, payload.Type)
		if insertErr != nil {
			logs.Error("failed to add update logs: %v", insertErr.Error())
		}

		if len(retryIds) > 0 && payload.TriesLeft > 0 {
			err := sendRetryMessage(payload.Type, payload.Realm, retryIds, payload.TriesLeft-1)
			if err != nil {
				return fmt.Errorf("failed to send retry message: %w", err)
			}
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to save snapshots: %w", err)
		}
		return scheduler.MarkComplete(payload)
	}

	return nil
}

func sendRetryMessage(taskType, realm string, ids []string, triesLeft int) error {
	// Create a new message to retry failed players
	var newPayload helpers.Payload
	newPayload.Type = taskType
	newPayload.Realm = realm
	newPayload.PlayerIDs = ids
	newPayload.TriesLeft = triesLeft
	newPayload.IsProcessing = false // This is a new message, so it's not processing

	return scheduler.UpdateQueueItem(newPayload)
}
