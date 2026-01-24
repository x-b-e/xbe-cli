package cli

import "github.com/spf13/cobra"

var projectTransportPlanSegmentDriversCmd = &cobra.Command{
	Use:   "project-transport-plan-segment-drivers",
	Short: "Browse project transport plan segment drivers",
	Long: `Browse project transport plan segment drivers on the XBE platform.

Project transport plan segment drivers associate drivers with transport plan segments.

Commands:
  list  List project transport plan segment drivers
  show  Show project transport plan segment driver details`,
	Example: `  # List project transport plan segment drivers
  xbe view project-transport-plan-segment-drivers list --limit 10

  # Show a project transport plan segment driver
  xbe view project-transport-plan-segment-drivers show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanSegmentDriversCmd)
}
