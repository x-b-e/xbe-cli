package cli

import "github.com/spf13/cobra"

var doTimeSheetCostCodeAllocationsCmd = &cobra.Command{
	Use:   "time-sheet-cost-code-allocations",
	Short: "Manage time sheet cost code allocations",
	Long: `Create, update, and delete time sheet cost code allocations.

Time sheet cost code allocations split a time sheet's costs across one or more
cost codes for billing and reporting.

Commands:
  create  Create a time sheet cost code allocation
  update  Update a time sheet cost code allocation
  delete  Delete a time sheet cost code allocation`,
	Example: `  # Create a time sheet cost code allocation
  xbe do time-sheet-cost-code-allocations create \
    --time-sheet 123 \
    --details '[{"cost_code_id":"456","percentage":1}]'

  # Update a time sheet cost code allocation
  xbe do time-sheet-cost-code-allocations update 789 \
    --details '[{"cost_code_id":"456","percentage":1}]'

  # Delete a time sheet cost code allocation
  xbe do time-sheet-cost-code-allocations delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doTimeSheetCostCodeAllocationsCmd)
}
