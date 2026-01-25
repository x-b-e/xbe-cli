package cli

import "github.com/spf13/cobra"

var projectStatusChangesCmd = &cobra.Command{
	Use:     "project-status-changes",
	Aliases: []string{"project-status-change"},
	Short:   "Browse project status changes",
	Long: `Browse project status changes.

Project status changes record the status history of projects.

Commands:
  list    List project status changes with filtering and pagination
  show    Show project status change details`,
	Example: `  # List project status changes
  xbe view project-status-changes list

  # Filter by project
  xbe view project-status-changes list --project 123

  # Filter by status
  xbe view project-status-changes list --status active

  # Show a status change
  xbe view project-status-changes show 456`,
}

func init() {
	viewCmd.AddCommand(projectStatusChangesCmd)
}
