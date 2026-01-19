package cli

import "github.com/spf13/cobra"

var actionItemsCmd = &cobra.Command{
	Use:   "action-items",
	Short: "Browse and view action items",
	Long: `Browse and view action items on the XBE platform.

Action items represent trackable work such as tasks, bugs, features,
integrations, and other items that can be assigned to individuals or
organizations and linked to projects.

Each action item includes:
  - Status (editing, ready_for_work, in_progress, in_verification, complete, on_hold)
  - Kind (feature, integration, sombrero, bug_fix, change_management, data_seeding, training)
  - Responsible person or organization
  - Associated project and tracker

Commands:
  list    List action items with filtering and pagination`,
	Example: `  # List action items
  xbe view action-items list

  # Filter by status
  xbe view action-items list --status in_progress
  xbe view action-items list --status ready_for_work

  # Filter by kind
  xbe view action-items list --kind bug_fix
  xbe view action-items list --kind feature

  # Filter by project
  xbe view action-items list --project 123

  # Get results as JSON
  xbe view action-items list --json`,
}

func init() {
	viewCmd.AddCommand(actionItemsCmd)
}
