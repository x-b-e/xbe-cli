package cli

import "github.com/spf13/cobra"

var timeSheetUnapprovalsCmd = &cobra.Command{
	Use:     "time-sheet-unapprovals",
	Aliases: []string{"time-sheet-unapproval"},
	Short:   "View time sheet unapprovals",
	Long: `View time sheet unapprovals.

Unapprovals record a status change from approved to submitted and
may include a comment.

Commands:
  list    List time sheet unapprovals
  show    Show time sheet unapproval details`,
	Example: `  # List unapprovals
  xbe view time-sheet-unapprovals list

  # Show an unapproval
  xbe view time-sheet-unapprovals show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetUnapprovalsCmd)
}
