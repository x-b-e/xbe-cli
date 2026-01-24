package cli

import "github.com/spf13/cobra"

var doCustomerTruckersCmd = &cobra.Command{
	Use:   "customer-truckers",
	Short: "Manage customer trucker links",
	Long: `Manage customer trucker links on the XBE platform.

Customer truckers link customers to approved truckers for a broker.

Commands:
  create    Create a customer trucker link
  delete    Delete a customer trucker link`,
	Example: `  # Create a customer trucker link
  xbe do customer-truckers create --customer 123 --trucker 456

  # Delete a customer trucker link
  xbe do customer-truckers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomerTruckersCmd)
}
