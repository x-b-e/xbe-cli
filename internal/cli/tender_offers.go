package cli

import "github.com/spf13/cobra"

var tenderOffersCmd = &cobra.Command{
	Use:     "tender-offers",
	Aliases: []string{"tender-offer"},
	Short:   "View tender offers",
	Long: `View tender offers.

Tender offers move tenders from editing to offered.

Commands:
  list    List tender offers
  show    Show tender offer details`,
	Example: `  # List tender offers
  xbe view tender-offers list

  # Show a tender offer
  xbe view tender-offers show 123

  # JSON output
  xbe view tender-offers list --json`,
}

func init() {
	viewCmd.AddCommand(tenderOffersCmd)
}
