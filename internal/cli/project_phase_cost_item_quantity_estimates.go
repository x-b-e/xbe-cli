package cli

import "github.com/spf13/cobra"

var projectPhaseCostItemQuantityEstimatesCmd = &cobra.Command{
	Use:   "project-phase-cost-item-quantity-estimates",
	Short: "View project phase cost item quantity estimates",
	Long: `View project phase cost item quantity estimates on the XBE platform.

Project phase cost item quantity estimates capture probabilistic estimates of
cost item quantities for a given estimate set.

Commands:
  list    List quantity estimates
  show    Show quantity estimate details`,
	Example: `  # List quantity estimates
  xbe view project-phase-cost-item-quantity-estimates list

  # Filter by project phase cost item
  xbe view project-phase-cost-item-quantity-estimates list --project-phase-cost-item 123

  # Show a quantity estimate
  xbe view project-phase-cost-item-quantity-estimates show 456`,
}

func init() {
	viewCmd.AddCommand(projectPhaseCostItemQuantityEstimatesCmd)
}
