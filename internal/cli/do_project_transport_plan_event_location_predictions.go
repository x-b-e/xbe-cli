package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanEventLocationPredictionsCmd = &cobra.Command{
	Use:     "project-transport-plan-event-location-predictions",
	Aliases: []string{"project-transport-plan-event-location-prediction"},
	Short:   "Manage project transport plan event location predictions",
	Long: `Manage project transport plan event location predictions.

Location predictions rank candidate transport locations for project transport
plan events. You can create predictions for an event or explicit event context,
update explicit context fields, or delete predictions.

Commands:
  create   Create a location prediction
  update   Update a location prediction
  delete   Delete a location prediction`,
	Example: `  # Create predictions for a project transport plan event
  xbe do project-transport-plan-event-location-predictions create \
    --project-transport-plan-event 123 \
    --transport-order 456

  # Update explicit context fields
  xbe do project-transport-plan-event-location-predictions update 789 \
    --strategy-set-id-explicit 321

  # Delete a prediction
  xbe do project-transport-plan-event-location-predictions delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanEventLocationPredictionsCmd)
}
