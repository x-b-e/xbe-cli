package cli

import "github.com/spf13/cobra"

var doBiddersCmd = &cobra.Command{
	Use:   "bidders",
	Short: "Manage bidders",
	Long: `Create, update, and delete bidders.

Bidders represent entities that submit bids within broker bidding workflows.
Bidders are scoped to a broker and may be marked as the broker's self bidder.

Commands:
  create    Create a new bidder
  update    Update an existing bidder
  delete    Delete a bidder`,
	Example: `  # Create a bidder
  xbe do bidders create --name "Acme Logistics" --broker 123 --is-self-for-broker false

  # Update a bidder
  xbe do bidders update 456 --name "Acme Logistics West"

  # Delete a bidder
  xbe do bidders delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doBiddersCmd)
}
