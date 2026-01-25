package cli

import "github.com/spf13/cobra"

var equipmentLocationEstimatesCmd = &cobra.Command{
	Use:   "equipment-location-estimates",
	Short: "View equipment location estimates",
	Long: `View equipment location estimates.

Equipment location estimates calculate the most recent known location for
specified equipment based on location events and movement stop completions.

Commands:
  list  List equipment location estimates`,
	Example: `  # Estimate location for a piece of equipment
  xbe view equipment-location-estimates list --equipment 123

  # Estimate location as of a specific time
  xbe view equipment-location-estimates list --equipment 123 --as-of 2026-01-23T12:00:00Z

  # Constrain the event window
  xbe view equipment-location-estimates list --equipment 123 \
    --earliest-event-at 2026-01-22T00:00:00Z \
    --latest-event-at 2026-01-23T00:00:00Z

  # Output as JSON
  xbe view equipment-location-estimates list --equipment 123 --json`,
}

func init() {
	viewCmd.AddCommand(equipmentLocationEstimatesCmd)
}
