package cli

import "github.com/spf13/cobra"

var materialTransactionCostCodeAllocationsCmd = &cobra.Command{
	Use:   "material-transaction-cost-code-allocations",
	Short: "Browse material transaction cost code allocations",
	Long: `Browse material transaction cost code allocations on the XBE platform.

Material transaction cost code allocations capture how a material transaction's
costs are distributed across cost codes for job costing.

Commands:
  list    List material transaction cost code allocations
  show    Show material transaction cost code allocation details`,
	Example: `  # List allocations
  xbe view material-transaction-cost-code-allocations list

  # Show allocation details
  xbe view material-transaction-cost-code-allocations show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionCostCodeAllocationsCmd)
}
