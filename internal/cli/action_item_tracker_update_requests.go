package cli

import "github.com/spf13/cobra"

var actionItemTrackerUpdateRequestsCmd = &cobra.Command{
	Use:   "action-item-tracker-update-requests",
	Short: "Browse action item tracker update requests",
	Long: `Browse action item tracker update requests on the XBE platform.

Action item tracker update requests ask assignees to provide updates and
can be fulfilled with update notes.

Commands:
  list    List update requests
  show    Show update request details`,
	Example: `  # List update requests
  xbe view action-item-tracker-update-requests list

  # Show an update request
  xbe view action-item-tracker-update-requests show 123

  # Output as JSON
  xbe view action-item-tracker-update-requests list --json`,
}

func init() {
	viewCmd.AddCommand(actionItemTrackerUpdateRequestsCmd)
}
