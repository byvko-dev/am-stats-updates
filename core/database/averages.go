package database

import (
	"context"
	"time"

	"github.com/byvko-dev/am-core/mongodb/driver"
	"github.com/byvko-dev/am-stats-updates/core/blitzstars"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionAverages = "vehicle-averages"

func GetTankAverages(tankId int) (blitzstars.TankAverages, error) {
	client, err := driver.NewClient()
	if err != nil {
		return blitzstars.TankAverages{}, err
	}

	var data blitzstars.TankAverages
	filter := make(map[string]interface{})
	filter["tankId"] = tankId
	return data, client.GetDocumentWithFilter(collectionAverages, filter, &data)
}

func UpdateTanksAverages(data ...blitzstars.TankAverages) error {
	if len(data) == 0 {
		return nil
	}

	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	var models []mongo.WriteModel
	for _, tank := range data {
		model := mongo.NewUpdateOneModel()
		model.SetFilter(bson.M{"tankId": tank.TankID})
		model.SetUpdate(bson.M{"$set": tank})
		model.SetUpsert(true)
		models = append(models, model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	_, err = client.Raw(collectionAverages).BulkWrite(ctx, models)
	return err
}
