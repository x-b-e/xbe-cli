package cli

import "github.com/spf13/cobra"

var equipmentMovementRequirementLocationsCmd = &cobra.Command{
	Use:     "equipment-movement-requirement-locations",
	Aliases: []string{"equipment-movement-requirement-location"},
	Short:   "View equipment movement requirement locations",
	Long: `View equipment movement requirement locations.

Equipment movement requirement locations represent named latitude/longitude
points used as origins and destinations for equipment movement requirements.

Commands:
  list    List locations with filtering
  show    Show location details`,
}

func init() {
	viewCmd.AddCommand(equipmentMovementRequirementLocationsCmd)
}
