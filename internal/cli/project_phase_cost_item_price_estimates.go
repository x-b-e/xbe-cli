package cli

import "github.com/spf13/cobra"

var projectPhaseCostItemPriceEstimatesCmd = &cobra.Command{
	Use:   "project-phase-cost-item-price-estimates",
	Short: "Browse project phase cost item price estimates",
	Long: `Browse project phase cost item price estimates on the XBE platform.

Price estimates capture probabilistic cost estimates for project phase cost items
within a project estimate set.

Commands:
  list    List project phase cost item price estimates
  show    Show project phase cost item price estimate details`,
	Example: `  # List project phase cost item price estimates
  xbe view project-phase-cost-item-price-estimates list

  # Show a project phase cost item price estimate
  xbe view project-phase-cost-item-price-estimates show 123`,
}

func init() {
	viewCmd.AddCommand(projectPhaseCostItemPriceEstimatesCmd)
}
