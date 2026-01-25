package cli

import "github.com/spf13/cobra"

var lineupJobProductionPlansCmd = &cobra.Command{
	Use:     "lineup-job-production-plans",
	Aliases: []string{"lineup-job-production-plan"},
	Short:   "Browse lineup job production plans",
	Long: `Browse lineup job production plans on the XBE platform.

Lineup job production plans link lineups to job production plans.

Commands:
  list    List lineup job production plans with filtering and pagination
  show    Show lineup job production plan details`,
	Example: `  # List lineup job production plans
  xbe view lineup-job-production-plans list

  # Show a lineup job production plan
  xbe view lineup-job-production-plans show 123

  # Output as JSON
  xbe view lineup-job-production-plans list --json`,
}

func init() {
	viewCmd.AddCommand(lineupJobProductionPlansCmd)
}
