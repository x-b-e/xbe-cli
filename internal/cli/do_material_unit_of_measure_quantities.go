package cli

import "github.com/spf13/cobra"

var doMaterialUnitOfMeasureQuantitiesCmd = &cobra.Command{
	Use:     "material-unit-of-measure-quantities",
	Aliases: []string{"material-unit-of-measure-quantity"},
	Short:   "Manage material unit of measure quantities",
	Long: `Create, update, and delete material unit of measure quantities.

Material unit of measure quantities track quantities recorded on material
transactions in a specific unit of measure.

Commands:
  create    Create a new material unit of measure quantity
  update    Update an existing material unit of measure quantity
  delete    Delete a material unit of measure quantity`,
}

func init() {
	doCmd.AddCommand(doMaterialUnitOfMeasureQuantitiesCmd)
}
