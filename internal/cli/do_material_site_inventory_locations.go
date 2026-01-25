package cli

import "github.com/spf13/cobra"

var doMaterialSiteInventoryLocationsCmd = &cobra.Command{
	Use:   "material-site-inventory-locations",
	Short: "Manage material site inventory locations",
	Long: `Create, update, and delete material site inventory locations.

Material site inventory locations represent specific stockpiles or inventory
areas within a material site.

Commands:
  create    Create a new material site inventory location
  update    Update an existing material site inventory location
  delete    Delete a material site inventory location`,
}

func init() {
	doCmd.AddCommand(doMaterialSiteInventoryLocationsCmd)
}
