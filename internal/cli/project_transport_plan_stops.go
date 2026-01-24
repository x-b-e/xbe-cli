package cli

import "github.com/spf13/cobra"

var projectTransportPlanStopsCmd = &cobra.Command{
	Use:   "project-transport-plan-stops",
	Short: "View project transport plan stops",
	Long: `View project transport plan stops.

Project transport plan stops represent ordered stops within a transport plan.`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanStopsCmd)
}
