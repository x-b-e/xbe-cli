package cli

import "github.com/spf13/cobra"

var biddersCmd = &cobra.Command{
	Use:   "bidders",
	Short: "View bidders",
	Long: `View bidders on the XBE platform.

Bidders represent organizations or entities that submit bids within a broker's
bidding workflows.

Commands:
  list    List bidders
  show    Show bidder details`,
	Example: `  # List bidders
  xbe view bidders list

  # Filter by broker
  xbe view bidders list --broker 123

  # Show a bidder
  xbe view bidders show 456`,
}

func init() {
	viewCmd.AddCommand(biddersCmd)
}
