package cli

import "github.com/spf13/cobra"

var doProjectTransportPlansCmd = &cobra.Command{
	Use:   "project-transport-plans",
	Short: "Manage project transport plans",
	Long: `Create, update, and delete project transport plans.

Project transport plans group transport orders and planned events,
segments, and assignments for moving materials.

Commands:
  create  Create a project transport plan
  update  Update a project transport plan
  delete  Delete a project transport plan`,
	Example: `  # Create a project transport plan
  xbe do project-transport-plans create --project 123

  # Update a project transport plan
  xbe do project-transport-plans update 456 --status approved

  # Delete a project transport plan
  xbe do project-transport-plans delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlansCmd)
}
