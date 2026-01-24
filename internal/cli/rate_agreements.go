package cli

import "github.com/spf13/cobra"

var rateAgreementsCmd = &cobra.Command{
	Use:     "rate-agreements",
	Aliases: []string{"rate-agreement"},
	Short:   "View rate agreements",
	Long:    "Commands for viewing rate agreements.",
}

func init() {
	viewCmd.AddCommand(rateAgreementsCmd)
}
