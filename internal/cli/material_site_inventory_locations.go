package cli

import "github.com/spf13/cobra"

var materialSiteInventoryLocationsCmd = &cobra.Command{
	Use:   "material-site-inventory-locations",
	Short: "View material site inventory locations",
	Long: `View material site inventory locations on the XBE platform.

Material site inventory locations represent specific stockpiles or inventory
areas within a material site.

Commands:
  list    List material site inventory locations with filtering
  show    Show material site inventory location details`,
	Example: `  # List material site inventory locations
  xbe view material-site-inventory-locations list

  # Filter by material site
  xbe view material-site-inventory-locations list --material-site 123

  # Show a specific inventory location
  xbe view material-site-inventory-locations show 456`,
}

func init() {
	viewCmd.AddCommand(materialSiteInventoryLocationsCmd)
}
