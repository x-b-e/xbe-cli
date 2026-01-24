package cli

import "github.com/spf13/cobra"

var actionItemLineItemsCmd = &cobra.Command{
	Use:   "action-item-line-items",
	Short: "Browse and view action item line items",
	Long: `Browse and view action item line items.

Action item line items are sub-tasks associated with action items. Each line
item can have a title, status, due date, and an optional responsible person.

Commands:
  list    List action item line items with filtering and pagination
  show    Show action item line item details`,
	Example: `  # List action item line items
  xbe view action-item-line-items list

  # Show a specific line item
  xbe view action-item-line-items show 123`,
}

func init() {
	viewCmd.AddCommand(actionItemLineItemsCmd)
}
