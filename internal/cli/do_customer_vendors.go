package cli

import "github.com/spf13/cobra"

var doCustomerVendorsCmd = &cobra.Command{
	Use:   "customer-vendors",
	Short: "Manage customer vendors",
	Long: `Manage customer-vendor trading partner relationships.

Commands:
  create    Create a customer-vendor relationship
  update    Update a customer-vendor relationship
  delete    Delete a customer-vendor relationship`,
	Example: `  # Create a customer-vendor relationship
  xbe do customer-vendors create --customer 123 --vendor "Trucker|456"

  # Update the external accounting ID
  xbe do customer-vendors update 789 --external-accounting-customer-vendor-id "ACCT-42"

  # Delete a customer-vendor relationship (requires --confirm)
  xbe do customer-vendors delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomerVendorsCmd)
}
