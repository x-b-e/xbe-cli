package cli

import "github.com/spf13/cobra"

var doCustomersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Manage customers",
	Long: `Manage customers on the XBE platform.

Commands:
  create    Create a new customer
  update    Update an existing customer
  delete    Delete a customer`,
	Example: `  # Create a customer
  xbe do customers create --name "ABC Construction" --broker 123

  # Update a customer
  xbe do customers update 456 --name "New Name"

  # Delete a customer (requires --confirm)
  xbe do customers delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomersCmd)
}
