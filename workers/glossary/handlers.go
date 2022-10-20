package glossary

import "fmt"

func UpdateGlossary() error {
	err := UpdateTankAverages()
	if err != nil {
		return fmt.Errorf("failed to update tank averages: %w", err)
	}

	err = UpdateVehiclesGlossary()
	if err != nil {
		return fmt.Errorf("failed to update vehicles glossary: %w", err)
	}

	// TODO: Update achievements glossary

	return nil
}
