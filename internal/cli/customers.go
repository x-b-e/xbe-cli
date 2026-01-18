package cli

import "github.com/spf13/cobra"

var customersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Browse and view customers",
	Long: `Browse and view customers on the XBE platform.

Customers are companies that purchase materials and services.
Use the list command to find customer IDs for filtering posts by creator.

Commands:
  list    List customers with filtering and pagination`,
	Example: `  # List customers
  xbe view customers list

  # Search by company name
  xbe view customers list --name "Acme"

  # Filter by active status
  xbe view customers list --active

  # Get results as JSON
  xbe view customers list --json --limit 10`,
}

func init() {
	viewCmd.AddCommand(customersCmd)
}
