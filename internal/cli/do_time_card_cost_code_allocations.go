package cli

import "github.com/spf13/cobra"

var doTimeCardCostCodeAllocationsCmd = &cobra.Command{
	Use:     "time-card-cost-code-allocations",
	Aliases: []string{"time-card-cost-code-allocation"},
	Short:   "Manage time card cost code allocations",
	Long: `Create, update, and delete time card cost code allocations.

Allocations define how a time card is split across cost codes using
percentages that must total 100%.`,
}

func init() {
	doCmd.AddCommand(doTimeCardCostCodeAllocationsCmd)
}
