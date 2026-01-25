package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanPlannedEventTimeSchedulesCmd = &cobra.Command{
	Use:     "project-transport-plan-planned-event-time-schedules",
	Aliases: []string{"project-transport-plan-planned-event-time-schedule"},
	Short:   "Generate planned event time schedules",
	Long: `Generate project transport plan planned event time schedules.

Schedules compute planned event times and warnings based on a project transport
plan or explicit transport order event data.

Commands:
  create   Generate a planned event time schedule`,
	Example: `  # Generate a schedule for a project transport plan
  xbe do project-transport-plan-planned-event-time-schedules create \
    --project-transport-plan 123

  # Generate a schedule from explicit plan data
  xbe do project-transport-plan-planned-event-time-schedules create \
    --transport-order 456 \
    --plan-data '{"events":[{"location_id":1,"event_type_id":2}]}'`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanPlannedEventTimeSchedulesCmd)
}
