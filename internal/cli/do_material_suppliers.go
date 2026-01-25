package cli

import "github.com/spf13/cobra"

var doMaterialSuppliersCmd = &cobra.Command{
	Use:   "material-suppliers",
	Short: "Manage material suppliers",
	Long: `Create, update, and delete material suppliers.

Material suppliers are companies that provide materials like asphalt, concrete,
aggregates, etc.

Commands:
  create    Create a new material supplier
  update    Update an existing material supplier
  delete    Delete a material supplier`,
}

func init() {
	doCmd.AddCommand(doMaterialSuppliersCmd)
}
