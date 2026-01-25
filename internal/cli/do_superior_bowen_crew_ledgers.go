package cli

import "github.com/spf13/cobra"

var doSuperiorBowenCrewLedgersCmd = &cobra.Command{
	Use:     "superior-bowen-crew-ledgers",
	Aliases: []string{"superior-bowen-crew-ledger"},
	Short:   "Create Superior Bowen crew ledgers",
	Long: `Create Superior Bowen crew ledgers.

Superior Bowen crew ledgers summarize labor and equipment costs for a job
production plan.

Commands:
  create    Create a Superior Bowen crew ledger`,
	Example: `  # Create a Superior Bowen crew ledger
  xbe do superior-bowen-crew-ledgers create --job-production-plan 123`,
}

func init() {
	doCmd.AddCommand(doSuperiorBowenCrewLedgersCmd)
}
