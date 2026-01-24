package cli

import "github.com/spf13/cobra"

var doMaterialTransactionRateSummariesCmd = &cobra.Command{
	Use:   "material-transaction-rate-summaries",
	Short: "Generate material transaction rate summaries",
	Long: `Generate material transaction rate summaries.

Material transaction rate summaries return hourly tons for a material site,
optionally filtered by time range and material type hierarchy.

Commands:
  create    Generate a material transaction rate summary`,
	Example: `  # Generate hourly rate summary for a material site
  xbe do material-transaction-rate-summaries create --material-site 123 --start-at 2025-01-01T00:00:00Z --end-at 2025-01-02T00:00:00Z

  # Filter by material type hierarchy
  xbe do material-transaction-rate-summaries create --material-site 123 --material-type-hierarchies "aggregate,asphalt"

  # Output JSON
  xbe do material-transaction-rate-summaries create --material-site 123 --json`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionRateSummariesCmd)
}
