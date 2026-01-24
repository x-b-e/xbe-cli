package cli

import "github.com/spf13/cobra"

var doProfitImprovementsCmd = &cobra.Command{
	Use:     "profit-improvements",
	Aliases: []string{"profit-improvement"},
	Short:   "Manage profit improvements",
	Long:    "Create, update, and delete profit improvements.",
}

func init() {
	doCmd.AddCommand(doProfitImprovementsCmd)
}
