package cli

import "github.com/spf13/cobra"

var timeCardUnapprovalsCmd = &cobra.Command{
	Use:     "time-card-unapprovals",
	Aliases: []string{"time-card-unapproval"},
	Short:   "View time card unapprovals",
	Long: `View time card unapprovals.

Unapprovals record a status change from approved to submitted and
may include a comment.

Commands:
  list    List time card unapprovals
  show    Show time card unapproval details`,
	Example: `  # List unapprovals
  xbe view time-card-unapprovals list

  # Show an unapproval
  xbe view time-card-unapprovals show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardUnapprovalsCmd)
}
