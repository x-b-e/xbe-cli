package cli

import "github.com/spf13/cobra"

var projectTransportPlansCmd = &cobra.Command{
	Use:     "project-transport-plans",
	Aliases: []string{"project-transport-plan"},
	Short:   "View project transport plans",
	Long: `View project transport plans.

Project transport plans group transport orders and the planned events,
segments, and assignments for moving materials.

Commands:
  list  List project transport plans
  show  Show project transport plan details`,
	Example: `  # List project transport plans
  xbe view project-transport-plans list

  # Show a project transport plan
  xbe view project-transport-plans show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlansCmd)
}
