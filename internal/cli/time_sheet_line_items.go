package cli

import "github.com/spf13/cobra"

var timeSheetLineItemsCmd = &cobra.Command{
	Use:   "time-sheet-line-items",
	Short: "Browse time sheet line items",
	Long: `Browse time sheet line items on the XBE platform.

Time sheet line items capture work segments, classifications, and timing details
for time sheets.

Commands:
  list    List time sheet line items
  show    Show time sheet line item details`,
	Example: `  # List time sheet line items
  xbe view time-sheet-line-items list

  # Show a time sheet line item
  xbe view time-sheet-line-items show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetLineItemsCmd)
}
