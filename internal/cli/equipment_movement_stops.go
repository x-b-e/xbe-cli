package cli

import "github.com/spf13/cobra"

var equipmentMovementStopsCmd = &cobra.Command{
	Use:     "equipment-movement-stops",
	Aliases: []string{"equipment-movement-stop"},
	Short:   "View equipment movement stops",
	Long: `View equipment movement stops.

Equipment movement stops represent ordered locations within an equipment
movement trip. Stops belong to a trip and reference a requirement location.

Commands:
  list    List stops with filtering
  show    Show stop details`,
}

func init() {
	viewCmd.AddCommand(equipmentMovementStopsCmd)
}
