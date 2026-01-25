package cli

import "github.com/spf13/cobra"

var doActionItemTrackerUpdateRequestsCmd = &cobra.Command{
	Use:   "action-item-tracker-update-requests",
	Short: "Manage action item tracker update requests",
	Long:  "Commands for creating, updating, and deleting action item tracker update requests.",
}

func init() {
	doCmd.AddCommand(doActionItemTrackerUpdateRequestsCmd)
}
