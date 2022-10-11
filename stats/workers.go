package stats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-cloud-functions/messaging"
	"github.com/byvko-dev/am-cloud-functions/scheduler"
	snapshots "github.com/byvko-dev/am-cloud-functions/stats/save-snapshots"
	accounts "github.com/byvko-dev/am-cloud-functions/stats/update-accounts"
	"github.com/byvko-dev/am-core/logs"
)

func StartUpdateWorkers(client *messaging.Client, subscription string, timeout time.Duration) error {
	block := make(chan bool)

	var handlerWrapper = func(data []byte) error {
		return handler(client, data)
	}

	err := client.Subscribe(subscription, timeout, handlerWrapper, func() { block <- true })
	if err != nil {
		return err
	}

	// Wait for timeout or exit
	<-block
	return nil
}

func handler(client *messaging.Client, data []byte) error {
	var payload helpers.Payload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		// Returning an error from this function will cause the message to be redelivered, so we don't want to do that
		logs.Error("failed to unmarshal payload: %v", err.Error())
		return nil
	}

	switch payload.Type {
	case scheduler.TaskTypeAccountUpdate:
		result, retryIds, err := accounts.UpdateRealmPlayers(payload.Realm)

		insertErr := database.AddUpdateLogs(result, payload.Type)
		if insertErr != nil {
			logs.Error("failed to add update logs: %v", insertErr.Error())
		}

		if len(retryIds) > 0 && payload.TriesLeft > 0 {
			err := sendRetryMessage(client, payload.Type, payload.Realm, retryIds, payload.TriesLeft-1)
			if err != nil {
				return fmt.Errorf("failed to send retry message: %w", err)
			}
		}
		return err

	case scheduler.TaskTypeSnapshot:
		result, retryIds, err := snapshots.SaveRealmSnapshots(payload.Realm, false)

		insertErr := database.AddUpdateLogs(result, payload.Type)
		if insertErr != nil {
			logs.Error("failed to add update logs: %v", insertErr.Error())
		}

		if len(retryIds) > 0 && payload.TriesLeft > 0 {
			err := sendRetryMessage(client, payload.Type, payload.Realm, retryIds, payload.TriesLeft-1)
			if err != nil {
				return fmt.Errorf("failed to send retry message: %w", err)
			}
		}
		return err
	}

	return nil
}

func sendRetryMessage(client *messaging.Client, taskType, realm string, ids []string, triesLeft int) error {
	// Create a new message to retry failed players
	var newPayload helpers.Payload
	newPayload.Type = taskType
	newPayload.Realm = realm
	newPayload.PlayerIDs = ids
	newPayload.TriesLeft = triesLeft

	message, err := json.Marshal(newPayload)
	if err != nil {
		return err
	}
	_, err = client.Publish(message, nil)

	return err
}
