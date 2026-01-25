package cli

import "github.com/spf13/cobra"

var doTimeSheetLineItemsCmd = &cobra.Command{
	Use:   "time-sheet-line-items",
	Short: "Manage time sheet line items",
	Long: `Create, update, and delete time sheet line items.

Time sheet line items capture the time, classification, and cost information
recorded on a time sheet.

Commands:
  create  Create a time sheet line item
  update  Update a time sheet line item
  delete  Delete a time sheet line item`,
	Example: `  # Create a time sheet line item
  xbe do time-sheet-line-items create --time-sheet 123 --start-at 2025-01-01T08:00:00Z --end-at 2025-01-01T12:00:00Z

  # Update a time sheet line item
  xbe do time-sheet-line-items update 123 --break-minutes 30

  # Delete a time sheet line item
  xbe do time-sheet-line-items delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeSheetLineItemsCmd)
}
