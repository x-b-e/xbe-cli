package cli

import "github.com/spf13/cobra"

var doCustomerCertificationTypesCmd = &cobra.Command{
	Use:   "customer-certification-types",
	Short: "Manage customer certification types",
	Long: `Manage customer certification types on the XBE platform.

Customer certification types link customers to certification types they can
track or require.

Commands:
  create    Create a customer certification type
  update    Update a customer certification type
  delete    Delete a customer certification type`,
	Example: `  # Create a customer certification type
  xbe do customer-certification-types create --customer 123 --certification-type 456

  # Update a customer certification type
  xbe do customer-certification-types update 789 --customer 123 --certification-type 456

  # Delete a customer certification type (requires --confirm)
  xbe do customer-certification-types delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomerCertificationTypesCmd)
}
