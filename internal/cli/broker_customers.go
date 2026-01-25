package cli

import "github.com/spf13/cobra"

var brokerCustomersCmd = &cobra.Command{
	Use:   "broker-customers",
	Short: "Browse broker-customer relationships",
	Long: `Browse broker-customer relationships.

Broker customers represent trading partner links between a broker and a customer.
Use these commands to list, inspect, and manage broker-customer records.

Commands:
  list    List broker customers with filtering and pagination
  show    Show a broker customer by ID`,
	Example: `  # List broker-customer relationships
  xbe view broker-customers list

  # Filter by broker
  xbe view broker-customers list --broker 123

  # Show a broker-customer relationship
  xbe view broker-customers show 456`,
}

func init() {
	viewCmd.AddCommand(brokerCustomersCmd)
}
