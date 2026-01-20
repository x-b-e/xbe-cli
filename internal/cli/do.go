package cli

import "github.com/spf13/cobra"

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Create, update, and delete XBE resources",
	Long: `Create, update, and delete XBE resources.

The do command provides write access to XBE platform data. Unlike view commands,
these operations modify data and require authentication.

Resources:
  action-items                     Manage action items (tasks, bugs, features)
  glossary-terms                   Manage glossary term definitions
  memberships                      Manage user-organization memberships`,
	Example: `  # Create an action item
  xbe do action-items create --title "Fix production bug" --kind bug_fix

  # Update an action item's status
  xbe do action-items update 123 --status in_progress

  # Delete an action item
  xbe do action-items delete 123 --confirm

  # Update a glossary term
  xbe do glossary-terms update 123 --definition "New definition"

  # Delete a glossary term
  xbe do glossary-terms delete 123 --confirm

  # Create a membership
  xbe do memberships create --user 123 --organization Broker|4 --kind manager

  # Update a membership
  xbe do memberships update 686 --kind operations

  # Delete a membership
  xbe do memberships delete 686 --confirm`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(doCmd)
}
