package cli

import "github.com/spf13/cobra"

var doBusinessUnitMembershipsCmd = &cobra.Command{
	Use:     "business-unit-memberships",
	Aliases: []string{"business-unit-membership"},
	Short:   "Manage business unit memberships",
	Long: `Manage business unit memberships on the XBE platform.

Commands:
  create    Create a business unit membership
  update    Update a business unit membership
  delete    Delete a business unit membership`,
	Example: `  # Create a business unit membership
  xbe do business-unit-memberships create --business-unit 123 --membership 456

  # Update a business unit membership
  xbe do business-unit-memberships update 789 --kind technician

  # Delete a business unit membership (requires --confirm)
  xbe do business-unit-memberships delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBusinessUnitMembershipsCmd)
}
