package cli

import "github.com/spf13/cobra"

var projectTransportPlanTractorsCmd = &cobra.Command{
	Use:     "project-transport-plan-tractors",
	Aliases: []string{"project-transport-plan-tractor"},
	Short:   "View project transport plan tractors",
	Long:    "Commands for viewing project transport plan tractors.",
}

func init() {
	viewCmd.AddCommand(projectTransportPlanTractorsCmd)
}
