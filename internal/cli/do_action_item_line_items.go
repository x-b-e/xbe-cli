package cli

import "github.com/spf13/cobra"

var doActionItemLineItemsCmd = &cobra.Command{
	Use:   "action-item-line-items",
	Short: "Manage action item line items",
	Long: `Create, update, and delete action item line items.

Action item line items are sub-tasks associated with action items. They can
track smaller tasks with a title, status, due date, and responsible person.

Commands:
  create  Create a new action item line item
  update  Update an action item line item
  delete  Delete an action item line item`,
	Example: `  # Create a line item
  xbe do action-item-line-items create --action-item 123 --title "Review plan"

  # Update a line item
  xbe do action-item-line-items update 456 --status closed

  # Delete a line item
  xbe do action-item-line-items delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doActionItemLineItemsCmd)
}
