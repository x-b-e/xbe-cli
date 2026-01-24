package cli

import "github.com/spf13/cobra"

var doDeveloperMembershipsCmd = &cobra.Command{
	Use:     "developer-memberships",
	Aliases: []string{"developer-membership"},
	Short:   "Manage developer memberships",
	Long: `Manage developer memberships on the XBE platform.

Commands:
  create    Create a developer membership
  update    Update a developer membership
  delete    Delete a developer membership`,
	Example: `  # Create a developer membership
  xbe do developer-memberships create --user 123 --developer 456

  # Update a developer membership
  xbe do developer-memberships update 789 --kind manager

  # Delete a developer membership (requires --confirm)
  xbe do developer-memberships delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doDeveloperMembershipsCmd)
}
