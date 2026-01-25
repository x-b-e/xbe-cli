package cli

import "github.com/spf13/cobra"

var projectRevenueItemQuantityEstimatesCmd = &cobra.Command{
	Use:   "project-revenue-item-quantity-estimates",
	Short: "View project revenue item quantity estimates",
	Long: `View project revenue item quantity estimates on the XBE platform.

Project revenue item quantity estimates capture probabilistic estimates
of revenue item quantities for a given estimate set.

Commands:
  list    List quantity estimates
  show    Show quantity estimate details`,
	Example: `  # List quantity estimates
  xbe view project-revenue-item-quantity-estimates list

  # Filter by project revenue item
  xbe view project-revenue-item-quantity-estimates list --project-revenue-item 123

  # Show a quantity estimate
  xbe view project-revenue-item-quantity-estimates show 456`,
}

func init() {
	viewCmd.AddCommand(projectRevenueItemQuantityEstimatesCmd)
}
