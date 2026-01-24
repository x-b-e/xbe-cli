package cli

import "github.com/spf13/cobra"

var doBrokerMembershipsCmd = &cobra.Command{
	Use:     "broker-memberships",
	Aliases: []string{"broker-membership"},
	Short:   "Manage broker memberships",
	Long: `Manage broker memberships on the XBE platform.

Commands:
  create    Create a broker membership
  update    Update a broker membership
  delete    Delete a broker membership`,
	Example: `  # Create a broker membership
  xbe do broker-memberships create --user 123 --broker 456

  # Update a broker membership
  xbe do broker-memberships update 789 --kind manager

  # Delete a broker membership (requires --confirm)
  xbe do broker-memberships delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerMembershipsCmd)
}
