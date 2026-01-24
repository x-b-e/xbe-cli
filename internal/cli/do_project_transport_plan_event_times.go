package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanEventTimesCmd = &cobra.Command{
	Use:     "project-transport-plan-event-times",
	Aliases: []string{"project-transport-plan-event-time"},
	Short:   "Manage project transport plan event times",
	Long:    "Create, update, and delete project transport plan event times.",
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanEventTimesCmd)
}
