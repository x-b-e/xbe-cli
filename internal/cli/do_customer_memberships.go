package cli

import "github.com/spf13/cobra"

var doCustomerMembershipsCmd = &cobra.Command{
	Use:     "customer-memberships",
	Aliases: []string{"customer-membership"},
	Short:   "Manage customer memberships",
	Long: `Manage customer memberships on the XBE platform.

Customer memberships link users to customers. Use these commands to create,
update, and delete customer memberships.

Commands:
  create    Create a customer membership
  update    Update a customer membership
  delete    Delete a customer membership`,
	Example: `  # Create a customer membership
  xbe do customer-memberships create --user 123 --customer 456 --kind manager

  # Update a customer membership
  xbe do customer-memberships update 789 --title "Dispatch" --is-admin true

  # Delete a customer membership
  xbe do customer-memberships delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doCustomerMembershipsCmd)
}
