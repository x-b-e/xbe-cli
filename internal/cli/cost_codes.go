package cli

import "github.com/spf13/cobra"

var costCodesCmd = &cobra.Command{
	Use:   "cost-codes",
	Short: "View cost codes",
	Long: `View cost codes on the XBE platform.

Cost codes are used to categorize and track costs for billing and accounting
purposes. They can be associated with specific customers, truckers, or brokers.

Commands:
  list    List cost codes`,
	Example: `  # List cost codes
  xbe view cost-codes list

  # Search by code
  xbe view cost-codes list --code "MAT-001"

  # Filter by broker
  xbe view cost-codes list --broker 123`,
}

func init() {
	viewCmd.AddCommand(costCodesCmd)
}
