package cli

import "github.com/spf13/cobra"

var transportOrderProjectTransportPlanStrategySetPredictionsCmd = &cobra.Command{
	Use:     "transport-order-project-transport-plan-strategy-set-predictions",
	Aliases: []string{"transport-order-project-transport-plan-strategy-set-prediction"},
	Short:   "Browse transport order strategy set predictions",
	Long: `Browse transport order strategy set predictions.

These predictions rank project transport plan strategy sets for a transport order
and are used to guide planning decisions.

Commands:
  list    List transport order strategy set predictions
  show    Show transport order strategy set prediction details`,
	Example: `  # List transport order strategy set predictions
  xbe view transport-order-project-transport-plan-strategy-set-predictions list

  # Filter by transport order
  xbe view transport-order-project-transport-plan-strategy-set-predictions list --transport-order 123

  # Show prediction details
  xbe view transport-order-project-transport-plan-strategy-set-predictions show 456`,
}

func init() {
	viewCmd.AddCommand(transportOrderProjectTransportPlanStrategySetPredictionsCmd)
}
