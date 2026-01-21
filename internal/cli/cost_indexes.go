package cli

import "github.com/spf13/cobra"

var costIndexesCmd = &cobra.Command{
	Use:   "cost-indexes",
	Short: "View cost indexes",
	Long: `View cost indexes on the XBE platform.

Cost indexes define pricing indexes that can be used for rate adjustments.
They can be broker-specific or global.

Commands:
  list    List cost indexes`,
	Example: `  # List cost indexes
  xbe view cost-indexes list

  # Filter by broker
  xbe view cost-indexes list --broker 123

  # Output as JSON
  xbe view cost-indexes list --json`,
}

func init() {
	viewCmd.AddCommand(costIndexesCmd)
}
