package database

import (
	"context"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-core/mongodb/driver"
)

const collectionLogs = "update-logs"

type Log struct {
	Task                 string `json:"task"`
	helpers.UpdateResult `bson:",inline"`
	Timestamp            time.Time `bson:"timestamp"`
}

func AddUpdateLogs(logs []helpers.UpdateResult, taskType string) error {
	if len(logs) == 0 {
		return nil
	}

	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	var logsToInsert []interface{}
	for _, log := range logs {
		logsToInsert = append(logsToInsert, Log{
			Task:         taskType,
			UpdateResult: log,
			Timestamp:    time.Now(),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = client.Raw(collectionLogs).InsertMany(ctx, logsToInsert)
	return err
}
