package cli

import "github.com/spf13/cobra"

var maintenanceRequirementsCmd = &cobra.Command{
	Use:   "requirements",
	Short: "View maintenance requirements",
	Long: `View maintenance requirements.

Maintenance requirements are individual maintenance tasks that need to be
performed on equipment. They can be organized into requirement sets and
linked to equipment.

Commands:
  list    List maintenance requirements with filtering
  show    View detailed requirement information`,
	Example: `  # List all requirements
  xbe view maintenance requirements list

  # Filter by set
  xbe view maintenance requirements list --set-id 123

  # Filter by equipment
  xbe view maintenance requirements list --equipment-id 456

  # Filter by status
  xbe view maintenance requirements list --status pending,in_progress

  # Show only templates
  xbe view maintenance requirements list --templates

  # View requirement details
  xbe view maintenance requirements show 789`,
}

func init() {
	maintenanceCmd.AddCommand(maintenanceRequirementsCmd)
}
