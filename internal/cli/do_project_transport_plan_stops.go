package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanStopsCmd = &cobra.Command{
	Use:     "project-transport-plan-stops",
	Aliases: []string{"project-transport-plan-stop"},
	Short:   "Manage project transport plan stops",
	Long:    "Create, update, and delete project transport plan stops.",
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanStopsCmd)
}
