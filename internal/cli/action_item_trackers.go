package cli

import "github.com/spf13/cobra"

var actionItemTrackersCmd = &cobra.Command{
	Use:     "action-item-trackers",
	Aliases: []string{"action-item-tracker"},
	Short:   "Browse action item trackers",
	Long: `Browse action item trackers on the XBE platform.

Action item trackers capture execution status, priority, and ownership for action
items. Trackers can be assigned to development and customer success assignees and
include effort sizing metadata.

Commands:
  list    List action item trackers
  show    Show action item tracker details`,
	Example: `  # List action item trackers
  xbe view action-item-trackers list

  # Show a tracker
  xbe view action-item-trackers show 123

  # JSON output
  xbe view action-item-trackers list --json`,
}

func init() {
	viewCmd.AddCommand(actionItemTrackersCmd)
}
