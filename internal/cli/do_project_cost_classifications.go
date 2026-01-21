package cli

import "github.com/spf13/cobra"

var doProjectCostClassificationsCmd = &cobra.Command{
	Use:   "project-cost-classifications",
	Short: "Manage project cost classifications",
	Long: `Create, update, and delete project cost classifications.

Project cost classifications define the hierarchy of cost categories for projects.
They are broker-scoped and can have parent-child relationships.

Commands:
  create  Create a new project cost classification
  update  Update an existing project cost classification
  delete  Delete a project cost classification`,
	Example: `  # Create a project cost classification
  xbe do project-cost-classifications create --name "Labor" --broker 123

  # Create a child classification
  xbe do project-cost-classifications create --name "Skilled Labor" --broker 123 --parent 456

  # Update a classification
  xbe do project-cost-classifications update 789 --name "Updated Name"

  # Delete a classification
  xbe do project-cost-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectCostClassificationsCmd)
}
