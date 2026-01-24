package cli

import "github.com/spf13/cobra"

var projectTransportPlanEventsCmd = &cobra.Command{
	Use:   "project-transport-plan-events",
	Short: "Browse project transport plan events",
	Long: `Browse project transport plan events on the XBE platform.

Project transport plan events represent ordered events on a project transport plan,
such as arrivals or departures at transport locations.

Commands:
  list  List project transport plan events
  show  Show project transport plan event details`,
	Example: `  # List project transport plan events
  xbe view project-transport-plan-events list --limit 10

  # Show a project transport plan event
  xbe view project-transport-plan-events show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanEventsCmd)
}
