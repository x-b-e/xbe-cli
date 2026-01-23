package cli

import "github.com/spf13/cobra"

var lineupScenarioLineupsCmd = &cobra.Command{
	Use:   "lineup-scenario-lineups",
	Short: "Browse lineup scenario lineups",
	Long: `Browse lineup scenario lineups.

Lineup scenario lineups link lineups to lineup scenarios for scheduling windows.

Commands:
  list    List lineup scenario lineups with filtering and pagination
  show    Show lineup scenario lineup details`,
	Example: `  # List lineup scenario lineups
  xbe view lineup-scenario-lineups list

  # Filter by lineup scenario
  xbe view lineup-scenario-lineups list --lineup-scenario 123

  # Show details
  xbe view lineup-scenario-lineups show 456`,
}

func init() {
	viewCmd.AddCommand(lineupScenarioLineupsCmd)
}
