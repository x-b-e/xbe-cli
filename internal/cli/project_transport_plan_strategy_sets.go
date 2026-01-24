package cli

import "github.com/spf13/cobra"

var projectTransportPlanStrategySetsCmd = &cobra.Command{
	Use:     "project-transport-plan-strategy-sets",
	Aliases: []string{"project-transport-plan-strategy-set"},
	Short:   "View project transport plan strategy sets",
	Long: `View project transport plan strategy sets.

Strategy sets group transport plan strategies into reusable patterns that
can be applied to project transport plans.

Commands:
  list  List project transport plan strategy sets
  show  Show project transport plan strategy set details`,
	Example: `  # List project transport plan strategy sets
  xbe view project-transport-plan-strategy-sets list

  # Show a project transport plan strategy set
  xbe view project-transport-plan-strategy-sets show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanStrategySetsCmd)
}
