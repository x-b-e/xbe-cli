package cli

import "github.com/spf13/cobra"

var businessUnitMembershipsCmd = &cobra.Command{
	Use:     "business-unit-memberships",
	Aliases: []string{"bu-memberships"},
	Short:   "View business unit memberships",
	Long: `View business unit memberships.

Business unit memberships link a user's broker membership to specific
business units, defining their role and access level within each unit.

Roles:
  manager     Can manage equipment and requirements in the BU
  technician  Can work on assigned maintenance items
  general     Limited to directly assigned items only

Commands:
  list    List business unit memberships`,
	Example: `  # List your business unit memberships
  xbe view business-unit-memberships list

  # List BU memberships for a specific user
  xbe view business-unit-memberships list --user-id 5724

  # List BU memberships for a specific membership
  xbe view business-unit-memberships list --membership-id 7627`,
}

func init() {
	viewCmd.AddCommand(businessUnitMembershipsCmd)
}
