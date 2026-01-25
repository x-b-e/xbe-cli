package cli

import "github.com/spf13/cobra"

var doJobProductionPlanUnscrappagesCmd = &cobra.Command{
	Use:   "job-production-plan-unscrappages",
	Short: "Unscrap job production plans",
	Long: `Unscrap job production plans on the XBE platform.

Unscrappages transition job production plans from scrapped to approved status.
Only plans in scrapped status can be unscrapped.

Commands:
  create    Unscrap a job production plan`,
	Example: `  # Unscrap a job production plan
  xbe do job-production-plan-unscrappages create --job-production-plan 123 --comment "Restoring plan"

  # Unscrap and suppress notifications
  xbe do job-production-plan-unscrappages create --job-production-plan 123 --suppress-status-change-notifications

  # Skip required mix design validation
  xbe do job-production-plan-unscrappages create --job-production-plan 123 --skip-validate-required-mix-designs`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanUnscrappagesCmd)
}
