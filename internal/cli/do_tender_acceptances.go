package cli

import "github.com/spf13/cobra"

var doTenderAcceptancesCmd = &cobra.Command{
	Use:     "tender-acceptances",
	Aliases: []string{"tender-acceptance"},
	Short:   "Accept tenders",
	Long: `Accept tenders.

Tender acceptances mark broker or customer tenders as accepted and can optionally
reject specific tender job schedule shifts.

Commands:
  create    Accept a tender`,
}

func init() {
	doCmd.AddCommand(doTenderAcceptancesCmd)
}
