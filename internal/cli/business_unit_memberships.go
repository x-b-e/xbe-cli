package cli

import "github.com/spf13/cobra"

var businessUnitMembershipsCmd = &cobra.Command{
	Use:     "business-unit-memberships",
	Aliases: []string{"business-unit-membership"},
	Short:   "View business unit memberships",
	Long: `View business unit memberships on the XBE platform.

Business unit memberships associate broker memberships with specific business units.

Commands:
  list    List business unit memberships
  show    Show business unit membership details`,
	Example: `  # List business unit memberships
  xbe view business-unit-memberships list

  # Show a business unit membership
  xbe view business-unit-memberships show 123`,
}

func init() {
	viewCmd.AddCommand(businessUnitMembershipsCmd)
}
