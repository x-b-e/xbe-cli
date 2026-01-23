package cli

import "github.com/spf13/cobra"

var doJobProductionPlanChangeSetsCmd = &cobra.Command{
	Use:   "job-production-plan-change-sets",
	Short: "Manage job production plan change sets",
	Long: `Commands for creating, updating, and deleting job production plan change sets.

Note: Change sets are immutable after creation; update and delete requests
will be rejected by the API.`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanChangeSetsCmd)
}
