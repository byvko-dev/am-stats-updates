package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/byvko-dev/am-core/mongodb/driver"
	"github.com/byvko-dev/am-types/stats/v3"
	"go.mongodb.org/mongo-driver/bson"
)

const collectionAccounts = "accounts"

func UpsertAccountInfo(info stats.AccountInfo) error {
	client, err := driver.NewClient()
	if err != nil {
		return err
	}
	update := make(map[string]interface{})
	update["$set"] = info

	return client.UpdateDocumentWithFilter(collectionAccounts, bson.M{"account_id": info.AccountID}, update, true)
}

func GetRealmAccountIDs(realm string) ([]int, error) {
	client, err := driver.NewClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.Raw(collectionAccounts).Distinct(ctx, "account_id", bson.M{"realm": strings.ToUpper(realm)})
	if err != nil {
		return nil, err
	}

	var ids []int
	for _, idRaw := range result {
		id, ok := idRaw.(int32)
		if !ok {
			return nil, errors.New("failed to convert account id to int")
		}
		ids = append(ids, int(id))
	}

	return ids, err
}
