package cli

import "github.com/spf13/cobra"

var customerMembershipsCmd = &cobra.Command{
	Use:     "customer-memberships",
	Aliases: []string{"customer-membership"},
	Short:   "Browse customer memberships",
	Long: `Browse customer memberships on the XBE platform.

Customer memberships link users to customer organizations and define
roles, permissions, and notification preferences.

Commands:
  list    List customer memberships with filtering
  show    Show customer membership details`,
	Example: `  # List customer memberships
  xbe view customer-memberships list

  # Show a customer membership
  xbe view customer-memberships show 123`,
}

func init() {
	viewCmd.AddCommand(customerMembershipsCmd)
}
