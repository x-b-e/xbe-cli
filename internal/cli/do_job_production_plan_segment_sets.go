package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSegmentSetsCmd = &cobra.Command{
	Use:     "job-production-plan-segment-sets",
	Aliases: []string{"job-production-plan-segment-set"},
	Short:   "Manage job production plan segment sets",
	Long: `Manage job production plan segment sets.

Commands:
  create    Create a job production plan segment set
  update    Update a job production plan segment set
  delete    Delete a job production plan segment set`,
	Example: `  # Create a job production plan segment set
  xbe do job-production-plan-segment-sets create --job-production-plan 123 --name "AM shift"

  # Update a job production plan segment set
  xbe do job-production-plan-segment-sets update 456 --start-offset-minutes 15

  # Delete a job production plan segment set
  xbe do job-production-plan-segment-sets delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanSegmentSetsCmd)
}
