package cli

import "github.com/spf13/cobra"

var doTimeCardPreApprovalsCmd = &cobra.Command{
	Use:     "time-card-pre-approvals",
	Aliases: []string{"time-card-pre-approval"},
	Short:   "Manage time card pre-approvals",
	Long: `Create, update, and delete time card pre-approvals.

Commands:
  create    Create a time card pre-approval
  update    Update a time card pre-approval
  delete    Delete a time card pre-approval`,
}

func init() {
	doCmd.AddCommand(doTimeCardPreApprovalsCmd)
}
