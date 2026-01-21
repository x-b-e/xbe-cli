package cli

import "github.com/spf13/cobra"

var doProjectRevenueClassificationsCmd = &cobra.Command{
	Use:   "project-revenue-classifications",
	Short: "Manage project revenue classifications",
	Long: `Create, update, and delete project revenue classifications.

Project revenue classifications define the hierarchy of revenue categories for projects.
They are broker-scoped and can have parent-child relationships.

Commands:
  create  Create a new project revenue classification
  update  Update an existing project revenue classification
  delete  Delete a project revenue classification`,
	Example: `  # Create a project revenue classification
  xbe do project-revenue-classifications create --name "Sales" --broker 123

  # Create a child classification
  xbe do project-revenue-classifications create --name "Product Sales" --broker 123 --parent 456

  # Update a classification
  xbe do project-revenue-classifications update 789 --name "Updated Name"

  # Delete a classification
  xbe do project-revenue-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectRevenueClassificationsCmd)
}
