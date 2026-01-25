package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanDriversCmd = &cobra.Command{
	Use:   "project-transport-plan-drivers",
	Short: "Manage project transport plan drivers",
	Long: `Create, update, and delete project transport plan driver assignments.

Project transport plan drivers assign drivers to segment ranges within
project transport plans.

Commands:
  create  Create a new project transport plan driver assignment
  update  Update an existing project transport plan driver assignment
  delete  Delete a project transport plan driver assignment`,
	Example: `  # Create a project transport plan driver assignment
  xbe do project-transport-plan-drivers create \\
    --project-transport-plan 123 \\
    --segment-start 456 \\
    --segment-end 789

  # Update a project transport plan driver assignment
  xbe do project-transport-plan-drivers update 555 --status pending

  # Delete a project transport plan driver assignment
  xbe do project-transport-plan-drivers delete 555 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanDriversCmd)
}
