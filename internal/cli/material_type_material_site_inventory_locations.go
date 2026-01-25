package cli

import "github.com/spf13/cobra"

var materialTypeMaterialSiteInventoryLocationsCmd = &cobra.Command{
	Use:   "material-type-material-site-inventory-locations",
	Short: "View material type material site inventory locations",
	Long: `View material type material site inventory locations on the XBE platform.

Material type material site inventory locations associate supplier-specific
material types with inventory locations at a material site.

Commands:
  list    List material type material site inventory locations
  show    Show material type material site inventory location details`,
	Example: `  # List material type material site inventory locations
  xbe view material-type-material-site-inventory-locations list

  # Filter by material type
  xbe view material-type-material-site-inventory-locations list --material-type 123

  # Filter by material site inventory location
  xbe view material-type-material-site-inventory-locations list --material-site-inventory-location 456

  # Show a specific mapping
  xbe view material-type-material-site-inventory-locations show 789`,
}

func init() {
	viewCmd.AddCommand(materialTypeMaterialSiteInventoryLocationsCmd)
}
