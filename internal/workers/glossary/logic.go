package glossary

import (
	"fmt"
	"os"
	"time"

	"github.com/byvko-dev/am-core/stats/blitzstars/v1"
	"github.com/byvko-dev/am-stats-updates/internal/core/database"
	"github.com/byvko-dev/am-types/stats/v3"
	"github.com/byvko-dev/am-types/wargaming/v2/glossary"
	wg "github.com/cufee/am-wg-proxy-next/client"
)

// Tank glossary
func UpdateVehiclesGlossary() error {
	client := wg.NewClient(os.Getenv("WG_PROXY_HOST"), time.Second*60)
	defer client.Close()

	data := make(map[int]stats.VehicleInfo)
	for _, lang := range glossary.AllLanguages {
		vehicles, err := client.GetVehiclesGlossary(lang)
		if err != nil {
			return fmt.Errorf("failed to get vehicles glossary: %w", err)
		}
		for _, v := range vehicles {
			if _, ok := data[v.TankID]; !ok {
				names := make(map[string]string)
				names[lang] = v.Name
				data[v.TankID] = stats.VehicleInfo{
					TankID:    v.TankID,
					Name:      names,
					Nation:    v.Nation,
					Type:      v.Type,
					Tier:      v.Tier,
					IsPremium: v.IsPremium,
				}
				continue
			}
			data[v.TankID].Name[lang] = v.Name
		}
	}

	var updates []stats.VehicleInfo
	for _, v := range data {
		updates = append(updates, v)
	}
	return database.UpdateVehicleGlossary(updates...)
}

func UpdateAchievements() error {
	return nil
}

func UpdateTankAverages() error {
	data, err := blitzstars.GetTankAverages()
	if err != nil {
		return fmt.Errorf("failed to get tank averages: %w", err)
	}
	return database.UpdateTanksAverages(data...)
}
