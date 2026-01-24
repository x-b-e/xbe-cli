package cli

import "github.com/spf13/cobra"

var doProjectProjectCostClassificationsCmd = &cobra.Command{
	Use:   "project-project-cost-classifications",
	Short: "Manage project project cost classifications",
	Long: `Manage project project cost classifications on the XBE platform.

Project project cost classifications link a project to a project cost classification
and optionally override the classification name for that project.

Commands:
  create    Create a new project project cost classification
  update    Update an existing project project cost classification
  delete    Delete a project project cost classification`,
	Example: `  # Create a project project cost classification
  xbe do project-project-cost-classifications create --project 123 --project-cost-classification 456

  # Update name override
  xbe do project-project-cost-classifications update 789 --name-override "Custom Name"

  # Delete a project project cost classification (requires --confirm)
  xbe do project-project-cost-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectProjectCostClassificationsCmd)
}
