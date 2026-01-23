package cli

import "github.com/spf13/cobra"

var doMaterialTransactionTicketGeneratorsCmd = &cobra.Command{
	Use:   "material-transaction-ticket-generators",
	Short: "Manage material transaction ticket generators",
	Long: `Manage material transaction ticket generators on the XBE platform.

Ticket generators define ticket numbering format rules for a broker or material
supplier organization.

Commands:
  create    Create a material transaction ticket generator
  update    Update a material transaction ticket generator
  delete    Delete a material transaction ticket generator`,
	Example: `  # Create a ticket generator
  xbe do material-transaction-ticket-generators create \
    --format-rule "MTX-{sequence}" \
    --organization-type brokers \
    --organization-id 123

  # Update a ticket generator
  xbe do material-transaction-ticket-generators update 456 --format-rule "MTX-{sequence}-A"

  # Delete a ticket generator (requires --confirm)
  xbe do material-transaction-ticket-generators delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionTicketGeneratorsCmd)
}
