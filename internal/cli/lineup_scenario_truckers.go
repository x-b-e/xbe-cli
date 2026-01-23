package cli

import "github.com/spf13/cobra"

var lineupScenarioTruckersCmd = &cobra.Command{
	Use:   "lineup-scenario-truckers",
	Short: "Browse lineup scenario truckers",
	Long: `Browse lineup scenario truckers.

Lineup scenario truckers configure assignment limits and constraints for a trucker
within a lineup scenario.

Commands:
  list    List lineup scenario truckers with filtering and pagination
  show    Show lineup scenario trucker details`,
	Example: `  # List lineup scenario truckers
  xbe view lineup-scenario-truckers list

  # Filter by lineup scenario
  xbe view lineup-scenario-truckers list --lineup-scenario 123

  # Show details
  xbe view lineup-scenario-truckers show 456`,
}

func init() {
	viewCmd.AddCommand(lineupScenarioTruckersCmd)
}
