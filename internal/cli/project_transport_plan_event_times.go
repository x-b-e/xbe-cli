package cli

import "github.com/spf13/cobra"

var projectTransportPlanEventTimesCmd = &cobra.Command{
	Use:   "project-transport-plan-event-times",
	Short: "View project transport plan event times",
	Long: `View project transport plan event times.

Project transport plan event times capture planned, expected, actual, and modeled
timestamps for project transport plan events.

Commands:
  list  List project transport plan event times
  show  Show project transport plan event time details`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanEventTimesCmd)
}
