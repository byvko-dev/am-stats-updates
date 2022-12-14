package calculations

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	blitzstars "github.com/byvko-dev/am-core/stats/blitzstars/v1/types"
	wn8 "github.com/byvko-dev/am-core/stats/ratings/wn8/v1"
	"github.com/byvko-dev/am-types/stats/v3"
	"github.com/byvko-dev/am-types/wargaming/v2/accounts"
	"github.com/byvko-dev/am-types/wargaming/v2/statistics"
)

func AccountSnapshot(account accounts.CompleteProfile, accountAchievements statistics.AchievementsFrame, vehicles []statistics.VehicleStatsFrame, vehicleAchievements map[int]statistics.AchievementsFrame, vehiclesCutoffTime int, vehicleAverages map[int]blitzstars.TankAverages, glossaryData map[int]stats.VehicleInfo) (stats.AccountSnapshot, error) {
	if account.AccountID == 0 {
		return stats.AccountSnapshot{}, errors.New("invalid account id")
	}

	var snapshot stats.AccountSnapshot
	snapshot.AccountID = int64(account.AccountID)
	snapshot.CreatedAt = time.Now().Unix()

	snapshot.LastBattleTime = int(account.LastBattleTime)
	snapshot.TotalBattles = int(account.Statistics.All.Battles + account.Statistics.Rating.Battles)

	var ratingSnapshot stats.Frame
	ratingSnapshot.Total = statistics.StatsFrame(account.Statistics.Rating)
	snapshot.Stats.Rating = ratingSnapshot

	var regularSnapshot stats.Frame
	regularSnapshot.Total = statistics.StatsFrame(account.Statistics.All)

	// Add achievements
	regularSnapshot.Achievements = accountAchievements

	// Add vehicles
	var wg sync.WaitGroup
	vehicleStats := make(chan stats.VehicleStats, len(vehicles))

	regularSnapshot.Vehicles = make(map[int]stats.VehicleStats)

	var totalWN8 int32
	var totalBattles int32

	for _, vehicle := range vehicles {
		if vehicle.LastBattleTime < vehiclesCutoffTime || vehicle.Stats.Battles == 0 {
			continue
		}
		wg.Add(1)
		go func(vehicle statistics.VehicleStatsFrame) {
			defer wg.Done()
			averages, ok := vehicleAverages[vehicle.TankID]
			ratings := make(map[string]int)
			if ok {
				rating, unweighted := wn8.VehicleWN8(vehicle, averages)
				ratings[wn8.WN8] = rating
				ratings[wn8.WN8Unweighted] = unweighted

				// For career WN8 calculation
				atomic.AddInt32(&totalWN8, int32(unweighted))
				atomic.AddInt32(&totalBattles, int32(vehicle.Stats.Battles))
			}

			v := stats.VehicleStats{
				VehicleStatsFrame: vehicle,
				Ratings:           ratings,
				Achievements:      vehicleAchievements[vehicle.TankID],
			}

			if vehicleInfo, ok := glossaryData[vehicle.TankID]; ok {
				v.TankName = vehicleInfo.Name
				v.TankTier = vehicleInfo.Tier
			}

			vehicleStats <- v
		}(vehicle)
	}
	wg.Wait()
	close(vehicleStats)
	for vehicle := range vehicleStats {
		regularSnapshot.Vehicles[vehicle.TankID] = vehicle
	}

	// Calculate career WN8
	regularSnapshot.Ratings = make(map[string]int)
	if totalBattles > 0 {
		regularSnapshot.Ratings["wn8"] = int(totalWN8 / totalBattles)
	}

	snapshot.Stats.Regular = regularSnapshot

	return snapshot, nil
}
