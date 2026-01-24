package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanTractorsCmd = &cobra.Command{
	Use:     "project-transport-plan-tractors",
	Aliases: []string{"project-transport-plan-tractor"},
	Short:   "Manage project transport plan tractors",
	Long:    "Commands for creating, updating, and deleting project transport plan tractors.",
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanTractorsCmd)
}
