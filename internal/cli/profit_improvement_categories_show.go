package cli

import "github.com/spf13/cobra"

func newProfitImprovementCategoriesShowCmd() *cobra.Command {
	return newGenericShowCmd("profit-improvement-categories")
}

func init() {
	profitImprovementCategoriesCmd.AddCommand(newProfitImprovementCategoriesShowCmd())
}
