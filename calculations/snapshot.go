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

func AccountSnapshot(account accounts.CompleteProfile, accountAchievements statistics.AchievementsFrame, vehicles []statistics.VehicleStatsFrame, vehicleAchievements map[int]statistics.AchievementsFrame, getAverage func(int) (blitzstars.TankAverages, error)) (stats.AccountSnapshot, error) {
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
		wg.Add(1)
		go func(vehicle statistics.VehicleStatsFrame) {
			defer wg.Done()
			averages, err := getAverage(vehicle.TankID)
			ratings := make(map[string]int)
			if err == nil {
				rating, unweighted := wn8.VehicleWN8(vehicle, averages)
				ratings[wn8.WN8] = rating
				ratings[wn8.WN8Unweighted] = unweighted
			}
			vehicleStats <- stats.VehicleStats{
				VehicleStatsFrame: vehicle,
				Ratings:           ratings,
				Achievements:      vehicleAchievements[vehicle.TankID],
			}

			// For career WN8 calculation
			unweighted, ok := ratings[wn8.WN8Unweighted]
			if ok {
				atomic.AddInt32(&totalWN8, int32(unweighted))
				atomic.AddInt32(&totalBattles, int32(vehicle.Stats.Battles))
			}
		}(vehicle)
	}
	wg.Wait()
	close(vehicleStats)
	for vehicle := range vehicleStats {
		regularSnapshot.Vehicles[vehicle.TankID] = vehicle
	}

	// Calculate career WN8
	if totalBattles > 0 {
		regularSnapshot.Ratings = make(map[string]int)
		regularSnapshot.Ratings["wn8"] = int(totalWN8 / totalBattles)
	}

	snapshot.Stats.Regular = regularSnapshot

	return snapshot, nil
}
