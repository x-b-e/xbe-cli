package cli

import "github.com/spf13/cobra"

var businessUnitCustomersCmd = &cobra.Command{
	Use:   "business-unit-customers",
	Short: "Browse business unit customer links",
	Long: `Browse business unit customer links.

Business unit customers associate customers with specific business units.

Commands:
  list    List business unit customers with filtering and pagination
  show    Show business unit customer details`,
	Example: `  # List business unit customer links
  xbe view business-unit-customers list

  # Filter by business unit
  xbe view business-unit-customers list --business-unit 123

  # Filter by customer
  xbe view business-unit-customers list --customer 456

  # Show a business unit customer link
  xbe view business-unit-customers show 789`,
}

func init() {
	viewCmd.AddCommand(businessUnitCustomersCmd)
}
