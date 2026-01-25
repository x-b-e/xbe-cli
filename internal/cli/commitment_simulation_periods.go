package cli

import "github.com/spf13/cobra"

var commitmentSimulationPeriodsCmd = &cobra.Command{
	Use:     "commitment-simulation-periods",
	Aliases: []string{"commitment-simulation-period"},
	Short:   "View commitment simulation periods",
	Long: `View commitment simulation periods.

Commitment simulation periods represent the period-level output of a
commitment simulation run, including date/window, iterations, and tons.`,
}

func init() {
	viewCmd.AddCommand(commitmentSimulationPeriodsCmd)
}
