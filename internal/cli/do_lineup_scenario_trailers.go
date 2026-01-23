package cli

import "github.com/spf13/cobra"

var doLineupScenarioTrailersCmd = &cobra.Command{
	Use:     "lineup-scenario-trailers",
	Aliases: []string{"lineup-scenario-trailer"},
	Short:   "Manage lineup scenario trailers",
	Long: `Manage lineup scenario trailers.

Commands:
  create    Create a lineup scenario trailer
  update    Update a lineup scenario trailer
  delete    Delete a lineup scenario trailer`,
	Example: `  # Create a lineup scenario trailer
  xbe do lineup-scenario-trailers create --lineup-scenario-trucker 123 --trailer 456

  # Update last assigned date
  xbe do lineup-scenario-trailers update 789 --last-assigned-on 2024-01-01

  # Delete a lineup scenario trailer
  xbe do lineup-scenario-trailers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doLineupScenarioTrailersCmd)
}
