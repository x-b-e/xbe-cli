package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSegmentsCmd = &cobra.Command{
	Use:   "job-production-plan-segments",
	Short: "Manage job production plan segments",
	Long: `Create, update, and delete job production plan segments.

Segments define production quantities, cycle timing, and related material details
within a job production plan.

Commands:
  create    Create a job production plan segment
  update    Update a job production plan segment
  delete    Delete a job production plan segment`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanSegmentsCmd)
}
