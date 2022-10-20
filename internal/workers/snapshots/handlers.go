package snapshots

import (
	"errors"
	"strconv"

	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/internal/core/database"
	"github.com/byvko-dev/am-stats-updates/internal/core/helpers"
)

// SaveRealmSnapshots generates and saves snapshots for all players on a realm
func SaveRealmSnapshots(realm string, isManual bool) ([]helpers.UpdateResult, []string, error) {
	if realm == "" {
		return nil, nil, errors.New("missing realm")
	}

	// Get players on realm from database
	logs.Debug("Getting player ids for realm %s", realm)
	playerIDs, err := database.GetRealmAccountIDs(realm)
	if err != nil {
		return nil, nil, err
	}
	var ids []string
	for _, id := range playerIDs {
		ids = append(ids, strconv.Itoa(id))
	}

	logs.Debug("Updating %d players on realm %s", len(playerIDs), realm)
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
	logs.Info("Updated %d players on realm %s", len(result), realm)
	return result, retryIds, nil
}
