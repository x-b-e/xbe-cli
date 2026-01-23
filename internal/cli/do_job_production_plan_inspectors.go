package cli

import "github.com/spf13/cobra"

var doJobProductionPlanInspectorsCmd = &cobra.Command{
	Use:     "job-production-plan-inspectors",
	Aliases: []string{"job-production-plan-inspector"},
	Short:   "Manage job production plan inspectors",
	Long: `Manage job production plan inspectors.

Commands:
  create    Create a job production plan inspector
  delete    Delete a job production plan inspector`,
	Example: `  # Create a job production plan inspector
  xbe do job-production-plan-inspectors create --job-production-plan-id 123 --user 456

  # Delete a job production plan inspector
  xbe do job-production-plan-inspectors delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanInspectorsCmd)
}
