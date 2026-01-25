package cli

import "github.com/spf13/cobra"

func newCostCodesShowCmd() *cobra.Command {
	return newGenericShowCmd("cost-codes")
}

func init() {
	costCodesCmd.AddCommand(newCostCodesShowCmd())
}
