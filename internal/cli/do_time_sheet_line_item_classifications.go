package cli

import "github.com/spf13/cobra"

var doTimeSheetLineItemClassificationsCmd = &cobra.Command{
	Use:   "time-sheet-line-item-classifications",
	Short: "Manage time sheet line item classifications",
	Long: `Create, update, and delete time sheet line item classifications.

Time sheet line item classifications categorize line items on time sheets.

Note: Only admin users can create, update, or delete classifications.

Commands:
  create  Create a new classification
  update  Update an existing classification
  delete  Delete a classification`,
	Example: `  # Create a classification
  xbe do time-sheet-line-item-classifications create --name "Overtime"

  # Update a classification
  xbe do time-sheet-line-item-classifications update 123 --name "Updated Name"

  # Delete a classification
  xbe do time-sheet-line-item-classifications delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeSheetLineItemClassificationsCmd)
}
