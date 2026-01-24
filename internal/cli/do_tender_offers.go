package cli

import "github.com/spf13/cobra"

var doTenderOffersCmd = &cobra.Command{
	Use:     "tender-offers",
	Aliases: []string{"tender-offer"},
	Short:   "Offer tenders",
	Long: `Offer tenders.

Tender offers move tenders from editing to offered.

Commands:
  create    Offer a tender`,
}

func init() {
	doCmd.AddCommand(doTenderOffersCmd)
}
