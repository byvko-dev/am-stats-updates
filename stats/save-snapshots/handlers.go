package savesnapshots

import (
	"errors"
	"strconv"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
)

// SaveRealmSnapshots generates and saves snapshots for all players on a realm
func SaveRealmSnapshots(realm string, isManual bool) ([]helpers.UpdateResult, []string, error) {
	if realm == "" {
		return nil, nil, errors.New("missing realm")
	}

	// Get players on realm from database
	playerIDs, err := database.GetRealmAccountIDs(realm)
	if err != nil {
		return nil, nil, err
	}
	var ids []string
	for _, id := range playerIDs {
		ids = append(ids, strconv.Itoa(id))
	}

	return SaveAccountSnapshots(realm, ids, isManual)
}

// SaveAccountSnapshots generates and saves snapshots for all player ids passed in
func SaveAccountSnapshots(realm string, accountIDs []string, isManual bool) ([]helpers.UpdateResult, []string, error) {
	if realm == "" {
		return nil, nil, errors.New("missing realm")
	}
	if len(accountIDs) == 0 {
		return nil, nil, errors.New("missing account ids")
	}

	result, retryIds := savePlayerSnapshots(realm, accountIDs, isManual)
	return result, retryIds, nil
}
