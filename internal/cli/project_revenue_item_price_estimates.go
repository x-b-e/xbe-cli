package cli

import "github.com/spf13/cobra"

var projectRevenueItemPriceEstimatesCmd = &cobra.Command{
	Use:     "project-revenue-item-price-estimates",
	Aliases: []string{"project-revenue-item-price-estimate"},
	Short:   "Browse project revenue item price estimates",
	Long:    "Commands for viewing project revenue item price estimates.",
}

func init() {
	viewCmd.AddCommand(projectRevenueItemPriceEstimatesCmd)
}
