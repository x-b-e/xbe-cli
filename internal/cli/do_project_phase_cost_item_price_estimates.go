package cli

import "github.com/spf13/cobra"

var doProjectPhaseCostItemPriceEstimatesCmd = &cobra.Command{
	Use:     "project-phase-cost-item-price-estimates",
	Aliases: []string{"project-phase-cost-item-price-estimate"},
	Short:   "Manage project phase cost item price estimates",
	Long: `Manage project phase cost item price estimates on the XBE platform.

Price estimates capture probabilistic cost estimates for project phase cost items
within a project estimate set.

Commands:
  create  Create a project phase cost item price estimate
  update  Update a project phase cost item price estimate
  delete  Delete a project phase cost item price estimate`,
	Example: `  # Create a price estimate
  xbe do project-phase-cost-item-price-estimates create --project-phase-cost-item 123 --project-estimate-set 456 --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'`,
}

func init() {
	doCmd.AddCommand(doProjectPhaseCostItemPriceEstimatesCmd)
}
