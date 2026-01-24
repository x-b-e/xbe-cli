package cli

import "github.com/spf13/cobra"

var customerVendorsCmd = &cobra.Command{
	Use:   "customer-vendors",
	Short: "Browse customer-vendor relationships",
	Long: `Browse customer-vendor relationships.

Customer vendors represent trading partner links between a customer and a vendor
(trucker). Use these commands to list, inspect, and manage customer-vendor
records.

Commands:
  list    List customer vendors with filtering and pagination
  show    Show a customer vendor by ID`,
	Example: `  # List customer-vendor relationships
  xbe view customer-vendors list

  # Filter by customer
  xbe view customer-vendors list --customer 123

  # Filter by vendor via partner filter
  xbe view customer-vendors list --partner "Trucker|456"

  # Show a customer-vendor relationship
  xbe view customer-vendors show 789`,
}

func init() {
	viewCmd.AddCommand(customerVendorsCmd)
}
