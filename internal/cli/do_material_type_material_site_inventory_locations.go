package cli

import "github.com/spf13/cobra"

var doMaterialTypeMaterialSiteInventoryLocationsCmd = &cobra.Command{
	Use:   "material-type-material-site-inventory-locations",
	Short: "Manage material type material site inventory locations",
	Long: `Create, update, and delete material type material site inventory locations.

Material type material site inventory locations associate supplier-specific
material types with inventory locations at a material site.

Commands:
  create    Create a new mapping
  update    Update an existing mapping
  delete    Delete a mapping`,
}

func init() {
	doCmd.AddCommand(doMaterialTypeMaterialSiteInventoryLocationsCmd)
}
