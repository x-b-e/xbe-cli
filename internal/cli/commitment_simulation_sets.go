package cli

import "github.com/spf13/cobra"

var commitmentSimulationSetsCmd = &cobra.Command{
	Use:     "commitment-simulation-sets",
	Aliases: []string{"commitment-simulation-set"},
	Short:   "View commitment simulation sets",
	Long: `View commitment simulation sets.

Commitment simulation sets define a simulation window and iteration count
for commitment planning, with processing status and linked simulations.`,
}

func init() {
	viewCmd.AddCommand(commitmentSimulationSetsCmd)
}
