package database

import (
	"context"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-core/mongodb/driver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const collectionQueue = "queue"

func AddQueueItems(items ...helpers.Payload) error {
	if len(items) == 0 {
		return nil
	}

	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	var insert []interface{}
	for _, item := range items {
		item.Timestamp = int(time.Now().Unix())
		insert = append(insert, item)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = client.Raw(collectionQueue).InsertMany(ctx, insert)
	return err
}

func GetNextTask(expiration time.Duration) (helpers.Payload, error) {
	client, err := driver.NewClient()
	if err != nil {
		return helpers.Payload{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"$or": []bson.M{
		{"triesLeft": bson.M{"$gt": 0}, "isProcessing": false, "timestamp": bson.M{"$gt": int(time.Now().Add(-expiration).Unix())}},
		{"triesLeft": bson.M{"$gt": 0}, "timestamp": bson.M{"$lt": int(time.Now().Add(-expiration * 2).Unix())}},
	}}

	var item helpers.Payload
	err = client.Raw(collectionQueue).FindOneAndUpdate(ctx, filter, bson.M{"$set": bson.M{"isProcessing": true}}).Decode(&item)
	return item, err
}

func UpdateQueueItem(item helpers.Payload) error {
	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Raw(collectionQueue).UpdateOne(ctx, bson.M{"_id": item.ID}, bson.M{"$set": item})
	return err
}

func DeleteQueueItem(id primitive.ObjectID) error {
	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Raw(collectionQueue).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func DeleteAllQueueItems() error {
	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Raw(collectionQueue).DeleteMany(ctx, bson.M{})
	return err
}
