package cli

import "github.com/spf13/cobra"

var doProjectRevenueItemPriceEstimatesCmd = &cobra.Command{
	Use:     "project-revenue-item-price-estimates",
	Aliases: []string{"project-revenue-item-price-estimate"},
	Short:   "Manage project revenue item price estimates",
	Long:    "Commands for creating, updating, and deleting project revenue item price estimates.",
}

func init() {
	doCmd.AddCommand(doProjectRevenueItemPriceEstimatesCmd)
}
