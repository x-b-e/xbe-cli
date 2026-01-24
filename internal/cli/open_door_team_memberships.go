package cli

import "github.com/spf13/cobra"

var openDoorTeamMembershipsCmd = &cobra.Command{
	Use:     "open-door-team-memberships",
	Aliases: []string{"open-door-team-membership"},
	Short:   "Browse open door team memberships",
	Long: `Browse open door team memberships.

Open door team memberships link memberships to organizations for Open Door issue
access.

Commands:
  list    List open door team memberships with filtering and pagination
  show    Show open door team membership details`,
	Example: `  # List open door team memberships
  xbe view open-door-team-memberships list

  # Show an open door team membership
  xbe view open-door-team-memberships show 123`,
}

func init() {
	viewCmd.AddCommand(openDoorTeamMembershipsCmd)
}
