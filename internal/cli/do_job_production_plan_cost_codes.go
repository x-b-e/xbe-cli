package cli

import "github.com/spf13/cobra"

var doJobProductionPlanCostCodesCmd = &cobra.Command{
	Use:     "job-production-plan-cost-codes",
	Aliases: []string{"job-production-plan-cost-code"},
	Short:   "Manage job production plan cost codes",
	Long: `Manage job production plan cost codes.

Commands:
  create    Create a job production plan cost code
  update    Update a job production plan cost code
  delete    Delete a job production plan cost code`,
	Example: `  # Create a job production plan cost code
  xbe do job-production-plan-cost-codes create --job-production-plan 123 --cost-code 456

  # Update a job production plan cost code
  xbe do job-production-plan-cost-codes update 789 --project-resource-classification 555

  # Delete a job production plan cost code
  xbe do job-production-plan-cost-codes delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanCostCodesCmd)
}
