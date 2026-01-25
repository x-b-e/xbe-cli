package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanSegmentDriversCmd = &cobra.Command{
	Use:   "project-transport-plan-segment-drivers",
	Short: "Manage project transport plan segment drivers",
	Long: `Create and delete project transport plan segment drivers.

Project transport plan segment drivers associate drivers with transport plan segments.

Commands:
  create  Create a project transport plan segment driver
  delete  Delete a project transport plan segment driver`,
	Example: `  # Create a project transport plan segment driver
  xbe do project-transport-plan-segment-drivers create \
    --project-transport-plan-segment 123 \
    --driver 456

  # Delete a project transport plan segment driver
  xbe do project-transport-plan-segment-drivers delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanSegmentDriversCmd)
}
