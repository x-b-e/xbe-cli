package cli

import "github.com/spf13/cobra"

var doCustomerTendersCmd = &cobra.Command{
	Use:     "customer-tenders",
	Aliases: []string{"customer-tender"},
	Short:   "Manage customer tenders",
	Long: `Manage customer tenders on the XBE platform.

Customer tenders are offers from customers to brokers for job work.

Commands:
  create  Create a customer tender
  update  Update a customer tender
  delete  Delete a customer tender`,
	Example: `  xbe do customer-tenders create --job 123 --customer 456 --broker 789
  xbe do customer-tenders update 123 --note "Updated note"
  xbe do customer-tenders delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomerTendersCmd)
}
