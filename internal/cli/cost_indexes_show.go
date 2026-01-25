package cli

import "github.com/spf13/cobra"

func newCostIndexesShowCmd() *cobra.Command {
	return newGenericShowCmd("cost-indexes")
}

func init() {
	costIndexesCmd.AddCommand(newCostIndexesShowCmd())
}
