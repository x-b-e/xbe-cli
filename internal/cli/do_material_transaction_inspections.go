package cli

import "github.com/spf13/cobra"

var doMaterialTransactionInspectionsCmd = &cobra.Command{
	Use:   "material-transaction-inspections",
	Short: "Manage material transaction inspections",
	Long: `Create, update, and delete material transaction inspections.

Material transaction inspections track inspection outcomes for material loads.

Commands:
  create    Create a material transaction inspection
  update    Update a material transaction inspection
  delete    Delete a material transaction inspection`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionInspectionsCmd)
}
