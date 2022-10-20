package database

import (
	"context"
	"time"

	"github.com/byvko-dev/am-core/mongodb/driver"
	"github.com/byvko-dev/am-core/stats/blitzstars/v1/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionAverages = "vehicle-averages"

func GetTankAverages(ids ...int) ([]types.TankAverages, error) {
	client, err := driver.NewClient()
	if err != nil {
		return nil, err
	}

	filter := bson.M{}
	filter["tankId"] = bson.M{"$in": ids}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var data []types.TankAverages
	cur, err := client.Raw(collectionAverages).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	return data, cur.All(ctx, &data)
}

func UpdateTanksAverages(data ...types.TankAverages) error {
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
