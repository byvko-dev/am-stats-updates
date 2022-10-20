package glossary

import (
	"fmt"
	"os"
	"time"

	"github.com/byvko-dev/am-stats-updates/core/blitzstars"
	"github.com/byvko-dev/am-stats-updates/core/database"
	"github.com/byvko-dev/am-types/wargaming/v2/glossary"
	wg "github.com/cufee/am-wg-proxy-next/client"
)

// Tank glossary
func UpdateVehiclesGlossary() error {
	client := wg.NewClient(os.Getenv("WG_PROXY_HOST"), time.Second*60)
	defer client.Close()

	data := make(map[int]glossary.VehicleDetails)
	for _, lang := range glossary.AllLanguages {
		vehicles, err := client.GetVehiclesGlossary(lang)
		if err != nil {
			return fmt.Errorf("failed to get vehicles glossary: %w", err)
		}
		for _, v := range vehicles {
			if _, ok := data[v.TankID]; !ok {
				data[v.TankID] = v
				continue
			}
			data[v.TankID].Name[lang] = v.Name[lang]
		}
	}

	var updates []glossary.VehicleDetails
	for _, v := range data {
		updates = append(updates, v)
	}
	return database.UpdateVehicleGlossary(updates...)
}

// Achievement glossary

func UpdateTankAverages() error {
	data, err := blitzstars.GetTankAverages()
	if err != nil {
		return fmt.Errorf("failed to get tank averages: %w", err)
	}
	return database.UpdateTanksAverages(data...)
}
