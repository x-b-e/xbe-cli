package cli

import "github.com/spf13/cobra"

var projectPhaseRevenueItemQuantityEstimatesCmd = &cobra.Command{
	Use:   "project-phase-revenue-item-quantity-estimates",
	Short: "View project phase revenue item quantity estimates",
	Long: `View project phase revenue item quantity estimates on the XBE platform.

Project phase revenue item quantity estimates capture probabilistic estimates
of revenue item quantities for a given estimate set.

Commands:
  list    List quantity estimates
  show    Show quantity estimate details`,
	Example: `  # List quantity estimates
  xbe view project-phase-revenue-item-quantity-estimates list

  # Filter by project phase revenue item
  xbe view project-phase-revenue-item-quantity-estimates list --project-phase-revenue-item 123

  # Show a quantity estimate
  xbe view project-phase-revenue-item-quantity-estimates show 456`,
}

func init() {
	viewCmd.AddCommand(projectPhaseRevenueItemQuantityEstimatesCmd)
}
