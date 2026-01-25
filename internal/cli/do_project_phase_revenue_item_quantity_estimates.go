package cli

import "github.com/spf13/cobra"

var doProjectPhaseRevenueItemQuantityEstimatesCmd = &cobra.Command{
	Use:   "project-phase-revenue-item-quantity-estimates",
	Short: "Manage project phase revenue item quantity estimates",
	Long: `Create, update, and delete project phase revenue item quantity estimates.

Project phase revenue item quantity estimates define probabilistic quantity inputs
for revenue items within a project estimate set.

Commands:
  create  Create a new quantity estimate
  update  Update an existing quantity estimate
  delete  Delete a quantity estimate`,
	Example: `  # Create a quantity estimate
  xbe do project-phase-revenue-item-quantity-estimates create \
    --project-phase-revenue-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

  # Update the estimate description
  xbe do project-phase-revenue-item-quantity-estimates update 789 --description "Updated estimate"

  # Delete a quantity estimate
  xbe do project-phase-revenue-item-quantity-estimates delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectPhaseRevenueItemQuantityEstimatesCmd)
}
