package cli

import "github.com/spf13/cobra"

var jobProductionPlanSupplyDemandBalancesCmd = &cobra.Command{
	Use:     "job-production-plan-supply-demand-balances",
	Aliases: []string{"job-production-plan-supply-demand-balance"},
	Short:   "Browse job production plan supply/demand balances",
	Long: `Browse job production plan supply/demand balances.

Supply/demand balances describe planned versus actual supply metrics for a job
production plan, including balance snapshots, trucks, and material transactions.

Commands:
  list    List supply/demand balances
  show    Show supply/demand balance details`,
	Example: `  # List supply/demand balances
  xbe view job-production-plan-supply-demand-balances list

  # Show a supply/demand balance
  xbe view job-production-plan-supply-demand-balances show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanSupplyDemandBalancesCmd)
}
