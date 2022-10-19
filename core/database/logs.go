package database

import (
	"context"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-core/mongodb/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionLogs = "update-logs"

type Log struct {
	Task                 string `json:"task"`
	helpers.UpdateResult `bson:",inline"`
	Timestamp            time.Time `bson:"timestamp"`
}

func UpsertUpdateLogs(logs []helpers.UpdateResult, taskType string) error {
	if len(logs) == 0 {
		return nil
	}

	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	var models []mongo.WriteModel
	for _, log := range logs {
		model := mongo.NewUpdateOneModel()
		model.SetFilter(bson.M{"accountID": log.AccountID, "task": taskType})
		model.SetUpdate(bson.M{"$set": Log{
			Task:         taskType,
			UpdateResult: log,
			Timestamp:    time.Now(),
		}})
		model.SetUpsert(true)
		models = append(models, model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = client.Raw(collectionLogs).BulkWrite(ctx, models)
	return err
}
