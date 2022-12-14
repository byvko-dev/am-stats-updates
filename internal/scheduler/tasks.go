package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/internal/core/database"
	"github.com/byvko-dev/am-stats-updates/internal/core/helpers"
)

const (
	TaskTypeSnapshot                   = "snapshotUpdate"
	TaskTypeAccountUpdate              = "accountUpdate"
	TaskTypeUpdateGlossaryVehicles     = "updateGlossaryVehicles"
	TaskTypeUpdateGlossaryAverages     = "updateGlossaryAverages"
	TaskTypeUpdateGlossaryAchievements = "updateGlossaryAchievements"
)

func CreateGlossaryTasks(taskType string, tries int) error {
	if taskType != TaskTypeUpdateGlossaryVehicles && taskType != TaskTypeUpdateGlossaryAchievements && taskType != TaskTypeUpdateGlossaryAverages {
		return errors.New("invalid task type")
	}
	logs.Info("Creating glossary tasks")

	var payload helpers.UpdateTask
	payload.Type = taskType
	payload.TriesLeft = tries

	err := AddQueueItem(payload)
	if err != nil {
		return fmt.Errorf("failed to add glossary task to queue: %w", err)
	}

	return nil
}

func CreateRealmTasks(taskType, realm string, tries, batchSize int) error {
	if taskType != TaskTypeSnapshot && taskType != TaskTypeAccountUpdate {
		return errors.New("invalid task type")
	}

	idsInt, err := database.GetRealmAccountIDs(realm)
	if err != nil {
		return err
	}

	logs.Debug("Creating tasks for realm %v with %v ids", realm, len(idsInt))

	if len(idsInt) == 0 {
		return nil
	}

	var ids []string
	for _, id := range idsInt {
		ids = append(ids, strconv.Itoa(id))
	}
	// Split the ids into chunks of batchSize
	chunks := make([][]string, 0, len(ids)/batchSize)
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[i:end])
	}

	for _, chunk := range chunks {
		var payload helpers.UpdateTask
		payload.Type = taskType
		payload.Realm = strings.ToUpper(realm)
		payload.PlayerIDs = chunk
		payload.TriesLeft = tries

		err := AddQueueItem(payload)
		if err != nil {
			return err
		}
	}
	return nil
}
