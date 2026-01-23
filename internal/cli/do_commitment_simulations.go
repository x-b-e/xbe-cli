package cli

import "github.com/spf13/cobra"

var doCommitmentSimulationsCmd = &cobra.Command{
	Use:   "commitment-simulations",
	Short: "Run commitment simulations",
	Long: `Run commitment simulations for commitments.

Commands:
  create    Create a commitment simulation`,
	Example: `  # Create a commitment simulation
  xbe do commitment-simulations create --commitment-type commitments --commitment-id 123 --start-on 2026-01-23 --end-on 2026-01-23 --iteration-count 100`,
}

func init() {
	doCmd.AddCommand(doCommitmentSimulationsCmd)
}
