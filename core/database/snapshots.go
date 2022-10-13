package database

import (
	"errors"
	"time"

	"github.com/byvko-dev/am-core/mongodb/driver"
	"github.com/byvko-dev/am-types/stats/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

const collectionSnapshots = "snapshots"

func SavePlayerSnapshot(snapshot stats.AccountSnapshot) error {
	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	_, err = client.InsertDocument(collectionSnapshots, snapshot)
	return err
}

func GetLastTotalBattles(accountId int, isManual bool) (int, error) {
	client, err := driver.NewClient()
	if err != nil {
		return 0, err
	}

	var target stats.AccountSnapshot
	filter := bson.M{"account_id": accountId, "is_manual": isManual}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var opts options.FindOneOptions
	opts.SetSort(bson.M{"created_at": -1})

	err = client.Raw(collectionSnapshots).FindOne(ctx, filter, &opts).Decode(&target)
	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		return 0, nil
	}
	return target.TotalBattles, err
}
