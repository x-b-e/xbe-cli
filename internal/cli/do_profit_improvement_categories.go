package cli

import "github.com/spf13/cobra"

var doProfitImprovementCategoriesCmd = &cobra.Command{
	Use:     "profit-improvement-categories",
	Aliases: []string{"profit-improvement-category"},
	Short:   "Manage profit improvement categories",
	Long:    `Create, update, and delete profit improvement categories.`,
}

func init() {
	doCmd.AddCommand(doProfitImprovementCategoriesCmd)
}
