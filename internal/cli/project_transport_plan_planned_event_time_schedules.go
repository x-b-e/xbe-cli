package cli

import "github.com/spf13/cobra"

var projectTransportPlanPlannedEventTimeSchedulesCmd = &cobra.Command{
	Use:     "project-transport-plan-planned-event-time-schedules",
	Aliases: []string{"project-transport-plan-planned-event-time-schedule"},
	Short:   "Browse project transport plan planned event time schedules",
	Long: `Browse project transport plan planned event time schedules.

Planned event time schedules generate a calculated schedule and warnings for
project transport plan events or explicit transport order event data.

Commands:
  list    List schedules with filtering and pagination
  show    Show full schedule details`,
	Example: `  # List schedules
  xbe view project-transport-plan-planned-event-time-schedules list

  # Filter by project transport plan
  xbe view project-transport-plan-planned-event-time-schedules list --project-transport-plan 123

  # Filter by transport order
  xbe view project-transport-plan-planned-event-time-schedules list --transport-order 456

  # Show schedule details
  xbe view project-transport-plan-planned-event-time-schedules show 789`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanPlannedEventTimeSchedulesCmd)
}
