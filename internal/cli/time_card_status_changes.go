package cli

import "github.com/spf13/cobra"

var timeCardStatusChangesCmd = &cobra.Command{
	Use:     "time-card-status-changes",
	Aliases: []string{"time-card-status-change"},
	Short:   "View time card status changes",
	Long: `View time card status changes.

Time card status changes capture status updates, who made the change,
when the change occurred, and any optional comments.

Commands:
  list    List time card status changes with filtering
  show    Show time card status change details`,
	Example: `  # List status changes
  xbe view time-card-status-changes list

  # Filter by time card
  xbe view time-card-status-changes list --time-card 123

  # Filter by status
  xbe view time-card-status-changes list --status submitted

  # Show a status change
  xbe view time-card-status-changes show 456`,
}

func init() {
	viewCmd.AddCommand(timeCardStatusChangesCmd)
}
