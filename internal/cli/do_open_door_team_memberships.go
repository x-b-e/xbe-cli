package cli

import "github.com/spf13/cobra"

var doOpenDoorTeamMembershipsCmd = &cobra.Command{
	Use:     "open-door-team-memberships",
	Aliases: []string{"open-door-team-membership"},
	Short:   "Manage open door team memberships",
	Long:    "Commands for creating, updating, and deleting open door team memberships.",
}

func init() {
	doCmd.AddCommand(doOpenDoorTeamMembershipsCmd)
}
