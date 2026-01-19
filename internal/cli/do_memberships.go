package cli

import "github.com/spf13/cobra"

var doMembershipsCmd = &cobra.Command{
	Use:   "memberships",
	Short: "Manage memberships",
	Long: `Manage memberships on the XBE platform.

Memberships define the relationship between users and organizations.
Use these commands to create, update, and delete memberships.

Commands:
  create    Create a new membership
  update    Update an existing membership
  delete    Delete a membership`,
	Example: `  # Create a membership
  xbe do memberships create --user 123 --organization Broker|4 --kind manager

  # Update a membership
  xbe do memberships update 686 --kind operations

  # Delete a membership (requires --confirm)
  xbe do memberships delete 686 --confirm`,
}

func init() {
	doCmd.AddCommand(doMembershipsCmd)
}
