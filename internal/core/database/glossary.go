package database

import (
	"context"
	"time"

	"github.com/byvko-dev/am-core/mongodb/driver"
	"github.com/byvko-dev/am-types/stats/v3"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionVehiclesGlossary = "vehicle-glossary"
const collectionAchievementsGlossary = "achievement-glossary"

func GetVehiclesInfo(ids ...int) ([]stats.VehicleInfo, error) {
	client, err := driver.NewClient()
	if err != nil {
		return nil, err
	}

	filter := bson.M{}
	filter["tank_id"] = bson.M{"$in": ids}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var data []stats.VehicleInfo
	cur, err := client.Raw(collectionVehiclesGlossary).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	return data, cur.All(ctx, &data)
}

func UpdateVehicleGlossary(data ...stats.VehicleInfo) error {
	if len(data) == 0 {
		return nil
	}

	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	var models []mongo.WriteModel
	for _, d := range data {
		model := mongo.NewUpdateOneModel()
		model.SetFilter(bson.M{"tank_id": d.TankID})
		model.SetUpdate(bson.M{"$set": d})
		model.SetUpsert(true)
		models = append(models, model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = client.Raw(collectionVehiclesGlossary).BulkWrite(ctx, models)
	return err
}

func GetAchievementInfo(achievementId int) (stats.AchievementInfo, error) {
	client, err := driver.NewClient()
	if err != nil {
		return stats.AchievementInfo{}, err
	}

	var data stats.AchievementInfo
	filter := make(map[string]interface{})
	filter["achievement_id"] = achievementId
	return data, client.GetDocumentWithFilter(collectionAchievementsGlossary, filter, &data)
}

func UpdateAchievementsGlossary(data ...stats.AchievementInfo) error {
	if len(data) == 0 {
		return nil
	}

	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	var models []mongo.WriteModel
	for _, d := range data {
		model := mongo.NewUpdateOneModel()
		model.SetFilter(bson.M{"achievement_id": d.ID})
		model.SetUpdate(bson.M{"$set": d})
		model.SetUpsert(true)
		models = append(models, model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = client.Raw(collectionAchievementsGlossary).BulkWrite(ctx, models)
	return err
}
