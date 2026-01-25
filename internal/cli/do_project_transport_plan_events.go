package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanEventsCmd = &cobra.Command{
	Use:   "project-transport-plan-events",
	Short: "Manage project transport plan events",
	Long: `Create, update, and delete project transport plan events.

Project transport plan events represent ordered events on a project transport plan,
such as arrivals or departures at transport locations.

Commands:
  create  Create a project transport plan event
  update  Update a project transport plan event
  delete  Delete a project transport plan event`,
	Example: `  # Create a project transport plan event
  xbe do project-transport-plan-events create \
    --project-transport-plan 123 \
    --project-transport-event-type 456

  # Update a project transport plan event
  xbe do project-transport-plan-events update 789 --external-tms-event-id "EVT-001"

  # Delete a project transport plan event
  xbe do project-transport-plan-events delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanEventsCmd)
}
