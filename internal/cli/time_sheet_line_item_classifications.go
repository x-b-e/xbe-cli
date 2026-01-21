package cli

import "github.com/spf13/cobra"

var timeSheetLineItemClassificationsCmd = &cobra.Command{
	Use:   "time-sheet-line-item-classifications",
	Short: "View time sheet line item classifications",
	Long: `View time sheet line item classifications on the XBE platform.

Time sheet line item classifications categorize line items on time sheets.

Commands:
  list    List time sheet line item classifications`,
	Example: `  # List time sheet line item classifications
  xbe view time-sheet-line-item-classifications list

  # Output as JSON
  xbe view time-sheet-line-item-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(timeSheetLineItemClassificationsCmd)
}
