package cli

import "github.com/spf13/cobra"

var timeSheetStatusChangesCmd = &cobra.Command{
	Use:     "time-sheet-status-changes",
	Aliases: []string{"time-sheet-status-change"},
	Short:   "View time sheet status changes",
	Long: `View time sheet status changes.

Status changes capture updates to a time sheet status, along with who made
the change, when it happened, and any comment.

Commands:
  list    List time sheet status changes
  show    Show time sheet status change details`,
	Example: `  # List time sheet status changes
  xbe view time-sheet-status-changes list

  # Show a time sheet status change
  xbe view time-sheet-status-changes show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetStatusChangesCmd)
}
