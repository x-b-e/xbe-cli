package cli

import "github.com/spf13/cobra"

var doTransportOrderProjectTransportPlanStrategySetPredictionsCmd = &cobra.Command{
	Use:     "transport-order-project-transport-plan-strategy-set-predictions",
	Aliases: []string{"transport-order-project-transport-plan-strategy-set-prediction"},
	Short:   "Manage transport order strategy set predictions",
	Long: `Create and manage transport order strategy set predictions.

Commands:
  create    Generate predictions for a transport order
  delete    Delete a prediction record`,
	Example: `  # Generate predictions for a transport order
  xbe do transport-order-project-transport-plan-strategy-set-predictions create --transport-order 123

  # Delete a prediction record (requires --confirm)
  xbe do transport-order-project-transport-plan-strategy-set-predictions delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doTransportOrderProjectTransportPlanStrategySetPredictionsCmd)
}
