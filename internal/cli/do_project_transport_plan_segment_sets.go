package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanSegmentSetsCmd = &cobra.Command{
	Use:   "project-transport-plan-segment-sets",
	Short: "Manage project transport plan segment sets",
	Long: `Create, update, and delete project transport plan segment sets.

Segment sets group segments within a project transport plan, optionally
assigned to a trucker.

Commands:
  create  Create a project transport plan segment set
  update  Update a project transport plan segment set
  delete  Delete a project transport plan segment set`,
	Example: `  # Create a project transport plan segment set
  xbe do project-transport-plan-segment-sets create --project-transport-plan 123

  # Update a project transport plan segment set
  xbe do project-transport-plan-segment-sets update 456 --trucker 789

  # Delete a project transport plan segment set
  xbe do project-transport-plan-segment-sets delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanSegmentSetsCmd)
}
