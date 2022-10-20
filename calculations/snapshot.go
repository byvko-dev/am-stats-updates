package calculations

import (
	"errors"
	"time"

	"github.com/byvko-dev/am-core/stats/blitzstars/v1"
	"github.com/byvko-dev/am-core/stats/ratings/wn8/v1"
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
	regularSnapshot.Vehicles = make(map[int]stats.VehicleStats)
	for _, vehicle := range vehicles {
		averages, err := getAverage(vehicle.TankID)
		ratings := make(map[string]int)
		if err == nil {
			rating, unweighted := wn8.VehicleWN8(vehicle, averages)
			ratings[wn8.WN8] = rating
			ratings[wn8.WN8Unweighted] = unweighted
		}
		regularSnapshot.Vehicles[vehicle.TankID] = stats.VehicleStats{
			VehicleStatsFrame: vehicle,
			Ratings:           ratings,
			Achievements:      vehicleAchievements[vehicle.TankID],
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

	return snapshot, nil
}
