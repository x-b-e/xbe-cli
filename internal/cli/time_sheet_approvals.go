package cli

import "github.com/spf13/cobra"

var timeSheetApprovalsCmd = &cobra.Command{
	Use:     "time-sheet-approvals",
	Aliases: []string{"time-sheet-approval"},
	Short:   "View time sheet approvals",
	Long: `View time sheet approvals.

Approvals record a status change from editing/submitted to approved and
may include a comment.

Commands:
  list    List time sheet approvals
  show    Show time sheet approval details`,
	Example: `  # List approvals
  xbe view time-sheet-approvals list

  # Show an approval
  xbe view time-sheet-approvals show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetApprovalsCmd)
}
