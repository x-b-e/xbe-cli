package cli

import "github.com/spf13/cobra"

var doMaterialTransactionCostCodeAllocationsCmd = &cobra.Command{
	Use:     "material-transaction-cost-code-allocations",
	Aliases: []string{"material-transaction-cost-code-allocation"},
	Short:   "Manage material transaction cost code allocations",
	Long:    "Commands for creating, updating, and deleting material transaction cost code allocations.",
}

func init() {
	doCmd.AddCommand(doMaterialTransactionCostCodeAllocationsCmd)
}
