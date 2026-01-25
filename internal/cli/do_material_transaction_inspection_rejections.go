package cli

import "github.com/spf13/cobra"

var doMaterialTransactionInspectionRejectionsCmd = &cobra.Command{
	Use:     "material-transaction-inspection-rejections",
	Aliases: []string{"material-transaction-inspection-rejection"},
	Short:   "Manage material transaction inspection rejections",
	Long: `Create, update, and delete material transaction inspection rejections.

Material transaction inspection rejections record rejected quantities and
notes for inspection results.

Commands:
  create    Create a new material transaction inspection rejection
  update    Update an existing material transaction inspection rejection
  delete    Delete a material transaction inspection rejection`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionInspectionRejectionsCmd)
}
