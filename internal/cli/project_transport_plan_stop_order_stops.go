package cli

import "github.com/spf13/cobra"

var projectTransportPlanStopOrderStopsCmd = &cobra.Command{
	Use:     "project-transport-plan-stop-order-stops",
	Aliases: []string{"project-transport-plan-stop-order-stop"},
	Short:   "View project transport plan stop order stops",
	Long: `View project transport plan stop order stops.

Project transport plan stop order stops link a project transport plan stop
to a transport order stop.

Commands:
  list  List project transport plan stop order stops
  show  Show project transport plan stop order stop details`,
	Example: `  # List project transport plan stop order stops
  xbe view project-transport-plan-stop-order-stops list

  # Show a project transport plan stop order stop
  xbe view project-transport-plan-stop-order-stops show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanStopOrderStopsCmd)
}
