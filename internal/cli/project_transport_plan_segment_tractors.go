package cli

import "github.com/spf13/cobra"

var projectTransportPlanSegmentTractorsCmd = &cobra.Command{
	Use:     "project-transport-plan-segment-tractors",
	Aliases: []string{"project-transport-plan-segment-tractor"},
	Short:   "View project transport plan segment tractors",
	Long:    "Commands for viewing project transport plan segment tractors.",
}

func init() {
	viewCmd.AddCommand(projectTransportPlanSegmentTractorsCmd)
}
