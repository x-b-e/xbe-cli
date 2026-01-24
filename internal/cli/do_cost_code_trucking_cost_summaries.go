package cli

import "github.com/spf13/cobra"

var doCostCodeTruckingCostSummariesCmd = &cobra.Command{
	Use:   "cost-code-trucking-cost-summaries",
	Short: "Manage cost code trucking cost summaries",
	Long: `Create and manage cost code trucking cost summaries.

Summaries are immutable once created. Updating an existing summary will
return an error.

Commands:
  create  Create a cost code trucking cost summary
  update  Update a cost code trucking cost summary
  delete  Delete a cost code trucking cost summary`,
	Example: `  # Create a summary
  xbe do cost-code-trucking-cost-summaries create --broker 123 --start-on 2025-01-01 --end-on 2025-01-31

  # Delete a summary
  xbe do cost-code-trucking-cost-summaries delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doCostCodeTruckingCostSummariesCmd)
}
