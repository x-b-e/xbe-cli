package cli

import "github.com/spf13/cobra"

var doActionItemTeamMembersCmd = &cobra.Command{
	Use:     "action-item-team-members",
	Aliases: []string{"action-item-team-member"},
	Short:   "Manage action item team members",
	Long:    "Commands for creating, updating, and deleting action item team members.",
}

func init() {
	doCmd.AddCommand(doActionItemTeamMembersCmd)
}
