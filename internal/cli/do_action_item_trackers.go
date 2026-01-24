package cli

import "github.com/spf13/cobra"

var doActionItemTrackersCmd = &cobra.Command{
	Use:     "action-item-trackers",
	Aliases: []string{"action-item-tracker"},
	Short:   "Manage action item trackers",
	Long: `Manage action item trackers on the XBE platform.

Commands:
  create    Create an action item tracker
  update    Update an action item tracker
  delete    Delete an action item tracker`,
	Example: `  # Create a tracker
  xbe do action-item-trackers create --action-item 123 --status ready_for_work

  # Update tracker status
  xbe do action-item-trackers update 123 --status in_development

  # Delete a tracker (requires --confirm)
  xbe do action-item-trackers delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doActionItemTrackersCmd)
}
