package cli

import "github.com/spf13/cobra"

var costCodeTruckingCostSummariesCmd = &cobra.Command{
	Use:   "cost-code-trucking-cost-summaries",
	Short: "View cost code trucking cost summaries",
	Long: `View cost code trucking cost summaries on the XBE platform.

Cost code trucking cost summaries aggregate approved time card costs by cost
code for a broker within a date range.

Commands:
  list    List cost code trucking cost summaries
  show    Show cost code trucking cost summary details`,
	Example: `  # List cost code trucking cost summaries
  xbe view cost-code-trucking-cost-summaries list

  # Filter by broker
  xbe view cost-code-trucking-cost-summaries list --broker 123

  # Show a summary with results
  xbe view cost-code-trucking-cost-summaries show 123

  # Output JSON
  xbe view cost-code-trucking-cost-summaries list --json`,
}

func init() {
	viewCmd.AddCommand(costCodeTruckingCostSummariesCmd)
}
