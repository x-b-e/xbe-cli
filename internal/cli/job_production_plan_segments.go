package cli

import "github.com/spf13/cobra"

var jobProductionPlanSegmentsCmd = &cobra.Command{
	Use:   "job-production-plan-segments",
	Short: "View job production plan segments",
	Long: `Browse job production plan segments on the XBE platform.

Job production plan segments describe the planned production breakdown for a job,
including quantities, cycle timing, and related material/site details.

Commands:
  list    List job production plan segments
  show    Show job production plan segment details`,
	Example: `  # List segments
  xbe view job-production-plan-segments list

  # Show a segment
  xbe view job-production-plan-segments show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanSegmentsCmd)
}
