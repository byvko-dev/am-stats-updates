package snapshots

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/byvko-dev/am-core/logs"
	wn8 "github.com/byvko-dev/am-core/stats/ratings/wn8/v1"
	"github.com/byvko-dev/am-stats-updates/core/database"
	"github.com/byvko-dev/am-stats-updates/core/helpers"
	"github.com/byvko-dev/am-types/stats/v3"
	"github.com/byvko-dev/am-types/wargaming/v2/accounts"
	"github.com/byvko-dev/am-types/wargaming/v2/statistics"
	wg "github.com/cufee/am-wg-proxy-next/client"
)

func savePlayerSnapshots(realm string, playerIDs []string, isManual bool) ([]helpers.UpdateResult, []string) {
	client := wg.NewClient(os.Getenv("WG_PROXY_HOST"), time.Second*60)
	defer client.Close()
	logs.Debug("Requesting %d accounts", len(playerIDs))
	accountData, err := client.BulkGetAccountsByID(playerIDs, realm)
	if err != nil {
		var result = make([]helpers.UpdateResult, len(playerIDs))
		for i, id := range playerIDs {
			result[i] = helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to get accounts: %s", err.Error()), WillRetry: true}
		}
		return result, playerIDs
	}
	if len(accountData) == 0 {
		logs.Error("No accounts returned")
		var result = make([]helpers.UpdateResult, len(playerIDs))
		for i, id := range playerIDs {
			result[i] = helpers.UpdateResult{AccountID: id, Error: "no accounts returned", WillRetry: true}
		}
		return result, playerIDs
	}

	logs.Debug("Requesting %v account clans", len(playerIDs))
	achievementsData, err := client.BulkGetAccountsAchievements(playerIDs, realm)
	if err != nil {
		var result = make([]helpers.UpdateResult, len(playerIDs))
		for i, id := range playerIDs {
			result[i] = helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to get achievements: %s", err.Error()), WillRetry: true}
		}
		return result, playerIDs
	}

	// Save all snapshots in goroutines
	var wg sync.WaitGroup
	var retry = make(chan string, len(accountData))
	var result = make(chan helpers.UpdateResult, len(accountData))
	for _, id := range playerIDs {
		wg.Add(1)
		account := accountData[id]
		go func(account accounts.CompleteProfile, id string) {
			defer wg.Done()
			if account.AccountID == 0 || id == "" {
				result <- helpers.UpdateResult{AccountID: id, Error: "account not found"}

				// This is a rare case where an account was deleted from WG servers
				idInt, err := strconv.Atoi(id)
				if err == nil {
					database.DeleteAccount(idInt)
				}
				return
			}

			{
				// TODO: This can be done in through aggregation pipeline in 1 query
				lastBattles, err := database.GetLastTotalBattles(int(account.AccountID), isManual)
				if err != nil {
					retry <- id
					result <- helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to get last total battles: %s", err.Error()), WillRetry: true}
					return
				}
				if (account.Statistics.All.Battles + account.Statistics.Rating.Battles - lastBattles) < 1 {
					// No new battles
					result <- helpers.UpdateResult{AccountID: id, Error: "no new battles", Success: true}
					return
				}
			}

			var snapshot stats.AccountSnapshot
			snapshot.AccountID = int64(account.AccountID)
			snapshot.CreatedAt = time.Now().Unix()
			snapshot.IsManual = isManual

			snapshot.LastBattleTime = int(account.LastBattleTime)
			snapshot.TotalBattles = int(account.Statistics.All.Battles + account.Statistics.Rating.Battles)

			var ratingSnapshot stats.SnapshotStats
			ratingSnapshot.Total = statistics.StatsFrame(account.Statistics.Rating)
			snapshot.Stats.Rating = ratingSnapshot

			var regularSnapshot stats.SnapshotStats
			regularSnapshot.Total = statistics.StatsFrame(account.Statistics.All)

			// Add achievements
			regularSnapshot.Achievements = achievementsData[id]

			// Get vehicle stats
			vehicles, e := client.GetAccountVehicles(int(account.AccountID))
			if e != nil {
				retry <- id
				result <- helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to get vehicles: %s", e.Message), WillRetry: true}
				return
			}

			// Get achievements per vehicle
			// TODO: no endpoint for this yet

			// Add vehicles
			regularSnapshot.Vehicles = make(map[int]stats.SnapshotVehicleStats)
			for _, vehicle := range vehicles {
				averages, err := database.GetTankAverages(vehicle.TankID)
				ratings := make(map[string]int)
				if err == nil {
					rating, unweighted := wn8.VehicleWN8(vehicle, averages)
					ratings[wn8.WN8] = rating
					ratings[wn8.WN8Unweighted] = unweighted
				}
				regularSnapshot.Vehicles[vehicle.TankID] = stats.SnapshotVehicleStats{
					VehicleStatsFrame: vehicle,
					Ratings:           ratings,
				}
			}

			snapshot.Stats.Regular = regularSnapshot

			// Career WN8
			var totalWN8 int
			var totalBattles int
			for _, vehicle := range regularSnapshot.Vehicles {
				unweighted, ok := vehicle.Ratings[wn8.WN8Unweighted]
				if ok {
					totalWN8 += unweighted
					totalBattles += vehicle.Stats.Battles
				}
			}
			if totalBattles > 0 {
				regularSnapshot.Ratings = make(map[string]int)
				regularSnapshot.Ratings["wn8"] = totalWN8 / totalBattles
			}

			// Save to database
			err := database.SavePlayerSnapshot(snapshot)
			if err != nil {
				retry <- id
				result <- helpers.UpdateResult{AccountID: id, Error: fmt.Sprintf("failed to save snapshot: %s", err.Error()), WillRetry: true}
			}

			result <- helpers.UpdateResult{AccountID: id, Success: true}
		}(account, id)
	}
	wg.Wait()
	close(result)
	close(retry)

	// Failed updates errors
	var results []helpers.UpdateResult
	for r := range result {
		results = append(results, r)
	}
	// Retry these IDs
	var retryIDs []string
	for id := range retry {
		retryIDs = append(retryIDs, id)
	}
	return results, retryIDs
}
