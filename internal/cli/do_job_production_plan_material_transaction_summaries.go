package cli

import "github.com/spf13/cobra"

var doJobProductionPlanMaterialTransactionSummariesCmd = &cobra.Command{
	Use:   "job-production-plan-material-transaction-summaries",
	Short: "Generate job production plan material transaction summaries",
	Long: `Generate job production plan material transaction summaries.

Summaries return accepted material transaction tons grouped by material type for a
single job production plan.

Commands:
  create    Generate a job production plan material transaction summary`,
	Example: `  # Summarize material transactions for a job production plan
  xbe do job-production-plan-material-transaction-summaries create --job-production-plan 123

  # Output JSON
  xbe do job-production-plan-material-transaction-summaries create --job-production-plan 123 --json`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanMaterialTransactionSummariesCmd)
}
