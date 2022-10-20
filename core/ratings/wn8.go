package ratings

import (
	"math"

	"github.com/byvko-dev/am-stats-updates/core/database"
	"github.com/byvko-dev/am-types/wargaming/v2/statistics"
)

// Calculates WN8 rating for a vehicle using cached tank averages. Returns rating and unweighted rating (raw)
func VehicleWN8(tank statistics.VehicleStatsFrame) (int, int) {
	// Get tank averages
	tankAvgData, err := database.GetTankAverages(tank.TankID)
	if err != nil {
		return 0, 0
	}
	battles := tank.Stats.Battles
	// Expected values for WN8
	expDef := tankAvgData.All.DroppedCapturePoints / tankAvgData.All.Battles
	expFrag := tankAvgData.Special.KillsPerBattle
	expSpot := tankAvgData.Special.SpotsPerBattle
	expDmg := tankAvgData.Special.DamagePerBattle
	expWr := tankAvgData.Special.Winrate

	// Actual performance
	pDef := float64(tank.Stats.DroppedCapturePoints) / float64(battles)
	pFrag := float64(tank.Stats.Frags) / float64(battles)
	pSpot := float64(tank.Stats.Spotted) / float64(battles)
	pDmg := float64(tank.Stats.DamageDealt) / float64(battles)
	pWr := float64(tank.Stats.Wins) / float64(battles) * 100

	// Calculate WN8 metrics
	rDef := pDef / expDef
	rFrag := pFrag / expFrag
	rSpot := pSpot / expSpot
	rDmg := pDmg / expDmg
	rWr := pWr / expWr

	adjustedWr := math.Max(0, ((rWr - 0.71) / (1 - 0.71)))
	adjustedDmg := math.Max(0, ((rDmg - 0.22) / (1 - 0.22)))
	adjustedDef := math.Max(0, (math.Min(adjustedDmg+0.1, (rDef-0.10)/(1-0.10))))
	adjustedSpot := math.Max(0, (math.Min(adjustedDmg+0.1, (rSpot-0.38)/(1-0.38))))
	adjustedFrag := math.Max(0, (math.Min(adjustedDmg+0.2, (rFrag-0.12)/(1-0.12))))

	rating := int(math.Round(((980 * adjustedDmg) + (210 * adjustedDmg * adjustedFrag) + (155 * adjustedFrag * adjustedSpot) + (75 * adjustedDef * adjustedFrag) + (145 * math.Min(1.8, adjustedWr)))))
	rawRating := rating * battles
	return rating, rawRating
}
