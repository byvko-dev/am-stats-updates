package glossary

import (
	"fmt"

	"github.com/byvko-dev/am-core/logs"
)

func UpdateGlossary() error {
	logs.Debug("Updating glossary")
	err := UpdateTankAverages()
	if err != nil {
		logs.Error("Failed to update tank averages: %s", err)
		return fmt.Errorf("failed to update tank averages: %w", err)
	}
	logs.Debug("Updated tank averages")

	err = UpdateVehiclesGlossary()
	if err != nil {
		logs.Error("Failed to update vehicles glossary: %s", err)
		return fmt.Errorf("failed to update vehicles glossary: %w", err)
	}
	logs.Debug("Updated vehicles glossary")

	// TODO: Update achievements glossary

	return nil
}
