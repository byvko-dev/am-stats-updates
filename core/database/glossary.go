package database

import (
	"context"
	"time"

	"github.com/byvko-dev/am-core/mongodb/driver"
	"github.com/byvko-dev/am-stats-updates/core/blitzstars"
	"github.com/byvko-dev/am-types/wargaming/v2/glossary"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionVehiclesGlossary = "vehicle-glossary"
const collectionAchievementsGlossary = "achievement-glossary"

func GetVehicleInfo(tankId int) (blitzstars.TankAverages, error) {
	client, err := driver.NewClient()
	if err != nil {
		return blitzstars.TankAverages{}, err
	}

	var data blitzstars.TankAverages
	filter := make(map[string]interface{})
	filter["tankId"] = tankId
	return data, client.GetDocumentWithFilter(collectionAverages, filter, &data)
}

func UpdateVehicleGlossary(data ...glossary.VehicleDetails) error {
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

func GetAchievementInfo(achievementId int) (glossary.AchievementDetails, error) {
	client, err := driver.NewClient()
	if err != nil {
		return glossary.AchievementDetails{}, err
	}

	var data glossary.AchievementDetails
	filter := make(map[string]interface{})
	filter["achievement_id"] = achievementId
	return data, client.GetDocumentWithFilter(collectionAchievementsGlossary, filter, &data)
}

func UpdateAchievementsGlossary(data ...glossary.AchievementDetails) error {
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
