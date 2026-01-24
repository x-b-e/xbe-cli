package cli

import "github.com/spf13/cobra"

var projectTransportPlanDriversCmd = &cobra.Command{
	Use:   "project-transport-plan-drivers",
	Short: "View project transport plan drivers",
	Long: `View project transport plan drivers.

Project transport plan drivers assign drivers to a segment range within
project transport plans.

Commands:
  list  List project transport plan driver assignments
  show  Show project transport plan driver details`,
	Example: `  # List project transport plan drivers
  xbe view project-transport-plan-drivers list

  # Show a project transport plan driver
  xbe view project-transport-plan-drivers show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanDriversCmd)
}
