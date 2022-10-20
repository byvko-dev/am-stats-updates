package blitzstars

import (
	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-stats-updates/core/helpers"
)

var apiURL = env.MustGetString("BLITZ_STARS_API_URI")

// GetTankAverages -
func GetTankAverages() (data []TankAverages, err error) {
	err = helpers.GetJSON(apiURL, &data)
	return data, err
}
