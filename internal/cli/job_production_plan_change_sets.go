package cli

import "github.com/spf13/cobra"

var jobProductionPlanChangeSetsCmd = &cobra.Command{
	Use:   "job-production-plan-change-sets",
	Short: "View job production plan change sets",
	Long: `View job production plan change sets on the XBE platform.

Change sets bundle scope filters and change instructions for job production
plans. Use list to browse change sets or show to inspect full details.

Commands:
  list    List change sets
  show    Show change set details`,
	Example: `  # List change sets
  xbe view job-production-plan-change-sets list

  # Show a change set
  xbe view job-production-plan-change-sets show 123

  # Output as JSON
  xbe view job-production-plan-change-sets list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanChangeSetsCmd)
}
