package cli

import "github.com/spf13/cobra"

var timeCardPreApprovalsCmd = &cobra.Command{
	Use:     "time-card-pre-approvals",
	Aliases: []string{"time-card-pre-approval"},
	Short:   "View time card pre-approvals",
	Long: `View time card pre-approvals.

Time card pre-approvals define approved maximum quantities and optional
automatic submission settings for a tender job schedule shift.

Commands:
  list    List time card pre-approvals
  show    Show time card pre-approval details`,
	Example: `  # List time card pre-approvals
  xbe view time-card-pre-approvals list

  # Show a time card pre-approval
  xbe view time-card-pre-approvals show 123

  # Output JSON
  xbe view time-card-pre-approvals list --json`,
}

func init() {
	viewCmd.AddCommand(timeCardPreApprovalsCmd)
}
