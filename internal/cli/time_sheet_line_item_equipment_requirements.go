package cli

import "github.com/spf13/cobra"

var timeSheetLineItemEquipmentRequirementsCmd = &cobra.Command{
	Use:   "time-sheet-line-item-equipment-requirements",
	Short: "View time sheet line item equipment requirements",
	Long: `View time sheet line item equipment requirements on the XBE platform.

Time sheet line item equipment requirements link equipment requirements to time sheet
line items, optionally marking a primary requirement.

Commands:
  list    List time sheet line item equipment requirements
  show    Show time sheet line item equipment requirement details`,
	Example: `  # List requirements
  xbe view time-sheet-line-item-equipment-requirements list

  # Show a requirement
  xbe view time-sheet-line-item-equipment-requirements show 123

  # JSON output
  xbe view time-sheet-line-item-equipment-requirements list --json`,
}

func init() {
	viewCmd.AddCommand(timeSheetLineItemEquipmentRequirementsCmd)
}
