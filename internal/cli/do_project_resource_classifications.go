package cli

import "github.com/spf13/cobra"

var doProjectResourceClassificationsCmd = &cobra.Command{
	Use:   "project-resource-classifications",
	Short: "Manage project resource classifications",
	Long: `Create, update, and delete project resource classifications.

Project resource classifications define categories for project resources.
They are broker-scoped and can have parent-child relationships.

Commands:
  create  Create a new project resource classification
  update  Update an existing project resource classification
  delete  Delete a project resource classification`,
	Example: `  # Create a project resource classification
  xbe do project-resource-classifications create --name "Equipment" --broker 123

  # Create a child classification
  xbe do project-resource-classifications create --name "Heavy Equipment" --broker 123 --parent 456

  # Update a classification
  xbe do project-resource-classifications update 789 --name "Updated Name"

  # Delete a classification
  xbe do project-resource-classifications delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectResourceClassificationsCmd)
}
