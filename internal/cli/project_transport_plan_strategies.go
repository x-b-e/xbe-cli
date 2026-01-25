package cli

import "github.com/spf13/cobra"

var projectTransportPlanStrategiesCmd = &cobra.Command{
	Use:     "project-transport-plan-strategies",
	Aliases: []string{"project-transport-plan-strategy"},
	Short:   "Browse project transport plan strategies",
	Long: `Browse project transport plan strategies.

Strategies define ordered steps and patterns used to plan transport events.

Commands:
  list    List strategies with filtering and pagination
  show    Show full details of a strategy`,
	Example: `  # List strategies
  xbe view project-transport-plan-strategies list

  # Filter by name
  xbe view project-transport-plan-strategies list --name "Default"

  # Filter by step pattern
  xbe view project-transport-plan-strategies list --step-pattern "pickup-dropoff"

  # Show strategy details
  xbe view project-transport-plan-strategies show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanStrategiesCmd)
}
