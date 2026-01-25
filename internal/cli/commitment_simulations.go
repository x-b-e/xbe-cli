package cli

import "github.com/spf13/cobra"

var commitmentSimulationsCmd = &cobra.Command{
	Use:   "commitment-simulations",
	Short: "Browse commitment simulations",
	Long: `Browse commitment simulations.

Commitment simulations model outcomes over a date range and iteration count for
an existing commitment.

Commands:
  list    List commitment simulations with filtering
  show    Show commitment simulation details`,
	Example: `  # List commitment simulations
  xbe view commitment-simulations list

  # Show a commitment simulation
  xbe view commitment-simulations show 123`,
}

func init() {
	viewCmd.AddCommand(commitmentSimulationsCmd)
}
