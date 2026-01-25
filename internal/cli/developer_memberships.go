package cli

import "github.com/spf13/cobra"

var developerMembershipsCmd = &cobra.Command{
	Use:     "developer-memberships",
	Aliases: []string{"developer-membership"},
	Short:   "View developer memberships",
	Long: `View developer memberships on the XBE platform.

Developer memberships define the relationship between users and developer organizations.

Commands:
  list    List developer memberships
  show    Show developer membership details`,
	Example: `  # List developer memberships
  xbe view developer-memberships list

  # Show a developer membership
  xbe view developer-memberships show 123`,
}

func init() {
	viewCmd.AddCommand(developerMembershipsCmd)
}
