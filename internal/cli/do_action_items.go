package cli

import "github.com/spf13/cobra"

var doActionItemsCmd = &cobra.Command{
	Use:   "action-items",
	Short: "Manage action items",
	Long: `Manage action items on the XBE platform.

Commands:
  create    Create a new action item
  update    Update an existing action item
  delete    Delete an action item (soft delete)`,
	Example: `  # Create an action item
  xbe do action-items create --title "Fix production bug" --kind bug_fix

  # Update an action item's status
  xbe do action-items update 123 --status in_progress

  # Delete an action item (requires --confirm)
  xbe do action-items delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doActionItemsCmd)
}
