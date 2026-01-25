package cli

import "github.com/spf13/cobra"

var doProjectPhaseCostItemQuantityEstimatesCmd = &cobra.Command{
	Use:   "project-phase-cost-item-quantity-estimates",
	Short: "Manage project phase cost item quantity estimates",
	Long: `Create, update, and delete project phase cost item quantity estimates.

Project phase cost item quantity estimates define probabilistic quantity inputs
for cost items within a project estimate set.

Commands:
  create  Create a new quantity estimate
  update  Update an existing quantity estimate
  delete  Delete a quantity estimate`,
	Example: `  # Create a quantity estimate
  xbe do project-phase-cost-item-quantity-estimates create \
    --project-phase-cost-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

  # Update the quantity basis
  xbe do project-phase-cost-item-quantity-estimates update 789 --revenue-item-quantity-basis 12.5

  # Delete a quantity estimate
  xbe do project-phase-cost-item-quantity-estimates delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectPhaseCostItemQuantityEstimatesCmd)
}
