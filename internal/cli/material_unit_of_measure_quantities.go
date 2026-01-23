package cli

import "github.com/spf13/cobra"

var materialUnitOfMeasureQuantitiesCmd = &cobra.Command{
	Use:     "material-unit-of-measure-quantities",
	Aliases: []string{"material-unit-of-measure-quantity"},
	Short:   "View material unit of measure quantities",
	Long: `View material unit of measure quantities.

Material unit of measure quantities track quantities recorded on material
transactions in a specific unit of measure.

Commands:
  list    List material unit of measure quantities with filtering
  show    Show material unit of measure quantity details`,
	Example: `  # List material unit of measure quantities
  xbe view material-unit-of-measure-quantities list

  # Show a material unit of measure quantity
  xbe view material-unit-of-measure-quantities show 123`,
}

func init() {
	viewCmd.AddCommand(materialUnitOfMeasureQuantitiesCmd)
}
