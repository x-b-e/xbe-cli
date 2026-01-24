package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanStopOrderStopsCmd = &cobra.Command{
	Use:   "project-transport-plan-stop-order-stops",
	Short: "Manage project transport plan stop order stops",
	Long:  "Commands for creating and deleting project transport plan stop order stops.",
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanStopOrderStopsCmd)
}
