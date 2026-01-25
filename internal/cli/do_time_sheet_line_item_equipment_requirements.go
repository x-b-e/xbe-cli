package cli

import "github.com/spf13/cobra"

var doTimeSheetLineItemEquipmentRequirementsCmd = &cobra.Command{
	Use:   "time-sheet-line-item-equipment-requirements",
	Short: "Manage time sheet line item equipment requirements",
	Long: `Create, update, and delete time sheet line item equipment requirements.

These links attach equipment requirements to time sheet line items and can
mark a requirement as primary for the line item.

Commands:
  create  Create a new requirement link
  update  Update an existing requirement link
  delete  Delete a requirement link`,
	Example: `  # Create a link
  xbe do time-sheet-line-item-equipment-requirements create \\
    --time-sheet-line-item 123 \\
    --equipment-requirement 456

  # Update a link
  xbe do time-sheet-line-item-equipment-requirements update 123 --is-primary true

  # Delete a link
  xbe do time-sheet-line-item-equipment-requirements delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeSheetLineItemEquipmentRequirementsCmd)
}
