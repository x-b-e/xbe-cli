package cli

import "github.com/spf13/cobra"

var doTenderReRatesCmd = &cobra.Command{
	Use:     "tender-re-rates",
	Aliases: []string{"tender-re-rate"},
	Short:   "Re-rate tenders",
	Long: `Re-rate tenders.

Tender re-rates re-run tender pricing and can optionally re-constrain rates.

Commands:
  create    Re-rate one or more tenders`,
}

func init() {
	doCmd.AddCommand(doTenderReRatesCmd)
}
