package database

import (
	"github.com/byvko-dev/am-core/mongodb/driver"
	"github.com/byvko-dev/am-types/stats/v3"
)

const collectionAccounts = "accounts"

func SaveAccountInfo(snapshot stats.AccountSnapshot) error {
	client, err := driver.NewClient()
	if err != nil {
		return err
	}

	_, err = client.InsertDocument(collectionSnapshots, snapshot)
	return err
}
