package cli

import "github.com/spf13/cobra"

var jobProductionPlanSegmentSetsCmd = &cobra.Command{
	Use:     "job-production-plan-segment-sets",
	Aliases: []string{"job-production-plan-segment-set"},
	Short:   "Browse job production plan segment sets",
	Long: `Browse job production plan segment sets on the XBE platform.

Job production plan segment sets group production plan segments and track offsets.

Commands:
  list    List job production plan segment sets with filtering and pagination
  show    Show job production plan segment set details`,
	Example: `  # List job production plan segment sets
  xbe view job-production-plan-segment-sets list

  # Show a job production plan segment set
  xbe view job-production-plan-segment-sets show 123

  # Output as JSON
  xbe view job-production-plan-segment-sets list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanSegmentSetsCmd)
}
