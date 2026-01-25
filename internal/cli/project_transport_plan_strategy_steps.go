package cli

import "github.com/spf13/cobra"

var projectTransportPlanStrategyStepsCmd = &cobra.Command{
	Use:     "project-transport-plan-strategy-steps",
	Aliases: []string{"project-transport-plan-strategy-step"},
	Short:   "Browse project transport plan strategy steps",
	Long: `Browse project transport plan strategy steps.

Project transport plan strategy steps define the ordered event types for a
transport plan strategy.

Commands:
  list  List project transport plan strategy steps
  show  Show project transport plan strategy step details`,
	Example: `  # List project transport plan strategy steps
  xbe view project-transport-plan-strategy-steps list --limit 10

  # Show a project transport plan strategy step
  xbe view project-transport-plan-strategy-steps show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanStrategyStepsCmd)
}
