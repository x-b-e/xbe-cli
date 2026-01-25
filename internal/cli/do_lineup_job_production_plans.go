package cli

import "github.com/spf13/cobra"

var doLineupJobProductionPlansCmd = &cobra.Command{
	Use:     "lineup-job-production-plans",
	Aliases: []string{"lineup-job-production-plan"},
	Short:   "Manage lineup job production plans",
	Long: `Manage lineup job production plans.

Commands:
  create    Create a lineup job production plan
  delete    Delete a lineup job production plan`,
	Example: `  # Create a lineup job production plan
  xbe do lineup-job-production-plans create --lineup 123 --job-production-plan 456

  # Delete a lineup job production plan
  xbe do lineup-job-production-plans delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doLineupJobProductionPlansCmd)
}
