package cli

import "github.com/spf13/cobra"

var truckerMembershipsCmd = &cobra.Command{
	Use:   "trucker-memberships",
	Short: "Browse trucker memberships",
	Long: `Browse and view trucker memberships on the XBE platform.

Trucker memberships link users to truckers and control role settings,
permissions, and notifications.

Commands:
  list    List trucker memberships with filtering and pagination
  show    Show trucker membership details`,
	Example: `  # List trucker memberships
  xbe view trucker-memberships list

  # Show a trucker membership
  xbe view trucker-memberships show 123`,
}

func init() {
	viewCmd.AddCommand(truckerMembershipsCmd)
}
