package cli

import "github.com/spf13/cobra"

var projectTransportPlanEventLocationPredictionsCmd = &cobra.Command{
	Use:     "project-transport-plan-event-location-predictions",
	Aliases: []string{"project-transport-plan-event-location-prediction"},
	Short:   "Browse project transport plan event location predictions",
	Long: `Browse project transport plan event location predictions.

Location predictions rank candidate transport locations for a project transport
plan event (or explicit event context) given a transport order.

Commands:
  list    List predictions with filtering and pagination
  show    Show full details of a prediction`,
	Example: `  # List predictions
  xbe view project-transport-plan-event-location-predictions list

  # Filter by project transport plan event
  xbe view project-transport-plan-event-location-predictions list --project-transport-plan-event 123

  # Filter by transport order
  xbe view project-transport-plan-event-location-predictions list --transport-order 456

  # Show prediction details
  xbe view project-transport-plan-event-location-predictions show 789`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanEventLocationPredictionsCmd)
}
