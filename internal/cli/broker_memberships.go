package cli

import "github.com/spf13/cobra"

var brokerMembershipsCmd = &cobra.Command{
	Use:     "broker-memberships",
	Aliases: []string{"broker-membership"},
	Short:   "View broker memberships",
	Long: `View broker memberships on the XBE platform.

Broker memberships define the relationship between users and broker organizations.

Commands:
  list    List broker memberships
  show    Show broker membership details`,
	Example: `  # List broker memberships
  xbe view broker-memberships list

  # Show a broker membership
  xbe view broker-memberships show 123`,
}

func init() {
	viewCmd.AddCommand(brokerMembershipsCmd)
}
