package cli

import "github.com/spf13/cobra"

var doCustomerRetainersCmd = &cobra.Command{
	Use:     "customer-retainers",
	Aliases: []string{"customer-retainer"},
	Short:   "Manage customer retainers",
	Long: `Create, update, and delete customer retainers.

Customer retainers define retainer agreements between a customer (buyer) and a
broker (seller).

Commands:
  create    Create a customer retainer
  update    Update a customer retainer
  delete    Delete a customer retainer`,
	Example: `  # Create a customer retainer
  xbe do customer-retainers create --customer 123 --broker 456 --status editing

  # Update a customer retainer
  xbe do customer-retainers update 789 --maximum-expected-daily-hours 10

  # Delete a customer retainer
  xbe do customer-retainers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomerRetainersCmd)
}
