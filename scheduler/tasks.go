package scheduler

import (
	"errors"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-cloud-functions/messaging"
)

const (
	TaskTypeSnapshot      = "snapshotUpdate"
	TaskTypeAccountUpdate = "accountUpdate"
)

func CreateRealmTasks(client *messaging.Client, taskType, realm string, tries int) error {
	if taskType != TaskTypeSnapshot && taskType != TaskTypeAccountUpdate {
		return errors.New("invalid task type")
	}

	idsInt, err := database.GetRealmAccountIDs(realm)
	if err != nil {
		return err
	}

	var ids []string
	for _, id := range idsInt {
		ids = append(ids, strconv.Itoa(id))
	}

	// Split the ids into chunks of 100
	chunks := make([][]string, 0, len(ids)/100)
	for i := 0; i < len(ids); i += 100 {
		end := i + 100
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[i:end])
	}

	for _, chunk := range chunks {
		var payload helpers.Payload
		payload.Type = taskType
		payload.Realm = strings.ToUpper(realm)
		payload.PlayerIDs = chunk
		payload.TriesLeft = tries

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		// Set attributes for message filtering
		attr := make(map[string]string)
		attr["type"] = taskType

		_, err = client.Publish(data, attr)
		if err != nil {
			return err
		}
	}
	return err
}
