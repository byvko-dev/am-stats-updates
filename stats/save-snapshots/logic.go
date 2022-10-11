package savesnapshots

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/byvko-dev/am-cloud-functions/core/database"
	"github.com/byvko-dev/am-types/stats/v3"
	"github.com/byvko-dev/am-types/wargaming/v1/accounts"
	"github.com/byvko-dev/am-types/wargaming/v2/statistics"
	wg "github.com/cufee/am-wg-proxy-next/client"
)

type FailedUpdate struct {
	AccountID string
	Err       error
}

func getRealmPlayers(realm string) ([]string, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func savePlayerSnapshots(realm string, playerIDs []string, isManual bool) []FailedUpdate {
	client := wg.NewClient(os.Getenv("WG_PROXY_HOST"), time.Second*30)
	accountData, err := client.BulkGetAccountsByID(playerIDs, realm)
	if err != nil {
		return []FailedUpdate{{AccountID: "all", Err: fmt.Errorf("failed to get accounts: %w", err)}}
	}
	achievementsData, err := client.BulkGetAccountsAchievements(playerIDs, realm)
	if err != nil {
		return []FailedUpdate{{AccountID: "all", Err: fmt.Errorf("failed to get achievements: %w", err)}}
	}

	// Save all snapshots in goroutines
	var wg sync.WaitGroup
	var failed = make(chan FailedUpdate, len(accountData))
	for _, id := range playerIDs {
		wg.Add(1)
		account := accountData[id]
		go func(account accounts.CompleteProfile, id string) {
			defer wg.Done()
			if account.AccountID == 0 || id == "" {
				// This should never happen but just in case
				failed <- FailedUpdate{AccountID: id, Err: errors.New("account not found")}
				return
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
				failed <- FailedUpdate{AccountID: id, Err: fmt.Errorf("failed to get vehicles: %s", e.Message)}
				return
			}

			// Get achievements per vehicle
			// TODO: no endpoint for this yet

			// Add vehicles
			regularSnapshot.Vehicles = make(map[int]stats.SnapshotVehicleStats)
			for _, vehicle := range vehicles {
				regularSnapshot.Vehicles[vehicle.TankID] = stats.SnapshotVehicleStats{
					VehicleStatsFrame: vehicle,
				}
			}

			snapshot.Stats.Regular = regularSnapshot

			// Save to database
			err := database.SavePlayerSnapshot(snapshot)
			if err != nil {
				failed <- FailedUpdate{AccountID: id, Err: fmt.Errorf("failed to save snapshot: %w", err)}
			}
		}(account, id)
	}
	wg.Wait()
	close(failed)

	var result []FailedUpdate
	for err := range failed {
		result = append(result, err)
	}
	return result
}
