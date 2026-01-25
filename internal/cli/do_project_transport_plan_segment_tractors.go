package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanSegmentTractorsCmd = &cobra.Command{
	Use:     "project-transport-plan-segment-tractors",
	Aliases: []string{"project-transport-plan-segment-tractor"},
	Short:   "Manage project transport plan segment tractors",
	Long:    "Commands for creating and deleting project transport plan segment tractors.",
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanSegmentTractorsCmd)
}
