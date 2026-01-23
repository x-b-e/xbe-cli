package cli

import "github.com/spf13/cobra"

var jobProductionPlanCostCodesCmd = &cobra.Command{
	Use:     "job-production-plan-cost-codes",
	Aliases: []string{"job-production-plan-cost-code"},
	Short:   "Browse job production plan cost codes",
	Long: `Browse job production plan cost codes on the XBE platform.

Job production plan cost codes map cost codes to job production plans.

Commands:
  list    List job production plan cost codes with filtering and pagination
  show    Show job production plan cost code details`,
	Example: `  # List job production plan cost codes
  xbe view job-production-plan-cost-codes list

  # Show a job production plan cost code
  xbe view job-production-plan-cost-codes show 123

  # Output as JSON
  xbe view job-production-plan-cost-codes list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanCostCodesCmd)
}
