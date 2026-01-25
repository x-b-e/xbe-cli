package cli

import "github.com/spf13/cobra"

var doTruckerMembershipsCmd = &cobra.Command{
	Use:   "trucker-memberships",
	Short: "Manage trucker memberships",
	Long: `Create, update, and delete trucker memberships.

Trucker memberships link users to truckers and control role settings,
permissions, and notifications.

Commands:
  create    Create a new trucker membership
  update    Update an existing trucker membership
  delete    Delete a trucker membership`,
}

func init() {
	doCmd.AddCommand(doTruckerMembershipsCmd)
}
