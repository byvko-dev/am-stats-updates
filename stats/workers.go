package stats

import (
	"fmt"
	"strconv"

	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/core/helpers"
	"github.com/byvko-dev/am-stats-updates/scheduler"
	snapshots "github.com/byvko-dev/am-stats-updates/stats/save-snapshots"
	accounts "github.com/byvko-dev/am-stats-updates/stats/update-accounts"
)

func StartUpdateWorkers(cancel chan int) error {
	workersNub := env.MustGetString("CONCURRENT_WORKERS")
	concurrency, _ := strconv.Atoi(workersNub)
	if concurrency < 1 {
		concurrency = 1
	}

	scheduler.SubscribeToTasks(concurrency, handler, cancel)
	return nil
}

func handler(payload helpers.UpdateTask) error {
	switch payload.Type {
	case scheduler.TaskTypeAccountUpdate:
		results, retryIds, err := accounts.UpdateSomePlayers(payload.Realm, payload.PlayerIDs)
		defer func() {
			err := scheduler.SaveUpdateResults(results, payload.Type) // Save results to DB
			if err != nil {
				logs.Error("Failed to save results: %v", err)
			}
		}()

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
		return nil

	case scheduler.TaskTypeSnapshot:
		results, retryIds, err := snapshots.SaveAccountSnapshots(payload.Realm, payload.PlayerIDs, false)
		defer func() {
			err := scheduler.SaveUpdateResults(results, payload.Type) // Save results to DB
			if err != nil {
				logs.Error("Failed to save results: %v", err)
			}
		}()

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
		return nil
	}

	return nil
}

func sendRetryMessage(taskType, realm string, ids []string, triesLeft int) error {
	// Create a new message to retry failed players
	var newPayload helpers.UpdateTask
	newPayload.Type = taskType
	newPayload.Realm = realm
	newPayload.PlayerIDs = ids
	newPayload.TriesLeft = triesLeft
	logs.Debug("Sending retry message for %v players", len(ids))
	return scheduler.AddQueueItem(newPayload)
}
