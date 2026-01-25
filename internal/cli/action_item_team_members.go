package cli

import "github.com/spf13/cobra"

var actionItemTeamMembersCmd = &cobra.Command{
	Use:     "action-item-team-members",
	Aliases: []string{"action-item-team-member"},
	Short:   "Browse action item team members",
	Long: `Browse action item team members.

Action item team members link users to action items and identify the
responsible person for the action item.

Commands:
  list    List action item team members with filtering and pagination
  show    Show full details of an action item team member`,
	Example: `  # List team members
  xbe view action-item-team-members list

  # Filter by action item
  xbe view action-item-team-members list --action-item 123

  # Filter by user
  xbe view action-item-team-members list --user 456

  # Show a team member
  xbe view action-item-team-members show 789`,
}

func init() {
	viewCmd.AddCommand(actionItemTeamMembersCmd)
}
