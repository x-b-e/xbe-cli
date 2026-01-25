package cli

import "github.com/spf13/cobra"

var timeSheetCostCodeAllocationsCmd = &cobra.Command{
	Use:   "time-sheet-cost-code-allocations",
	Short: "View time sheet cost code allocations",
	Long: `View time sheet cost code allocations.

Time sheet cost code allocations split a time sheet's costs across one or more
cost codes for billing and reporting.

Commands:
  list    List time sheet cost code allocations
  show    Show time sheet cost code allocation details`,
	Example: `  # List time sheet cost code allocations
  xbe view time-sheet-cost-code-allocations list

  # Show a specific allocation
  xbe view time-sheet-cost-code-allocations show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetCostCodeAllocationsCmd)
}
