package updateplayers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-cloud-functions/core/helpers"
	"github.com/byvko-dev/am-core/logs"
)

// UpdateRealmPlayers updates all players on a realm
func UpdateRealmPlayers(realm string) ([]helpers.UpdateResult, []string, error) {
	if realm == "" {
		return nil, nil, errors.New("missing realm")
	}

	logs.Debug("Getting player ids for realm %s", realm)
	ids, err := database.GetRealmAccountIDs(realm)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get realm accounts: %w", err)
	}

	idsStr := make([]string, 0, len(ids))
	for _, id := range ids {
		idsStr = append(idsStr, strconv.Itoa(id))
	}

	logs.Debug("Updating %d players on realm %s", len(idsStr), realm)
	return UpdateSomePlayers(realm, idsStr)
}

// UpdateSomePlayers updates all player ids passed in
func UpdateSomePlayers(realm string, accountIDs []string) ([]helpers.UpdateResult, []string, error) {
	if realm == "" {
		return nil, nil, errors.New("missing realm")
	}
	if len(accountIDs) == 0 {
		return nil, nil, errors.New("missing account ids")
	}

	result, retryIds := updateAccounts(realm, accountIDs)
	return result, retryIds, nil
}
