package cli

import "github.com/spf13/cobra"

var materialSiteReadingMaterialTypesCmd = &cobra.Command{
	Use:   "material-site-reading-material-types",
	Short: "View material site reading material types",
	Long: `View material site reading material types on the XBE platform.

Material site reading material types map external material identifiers from
plant or system readings to internal material types for a specific material site.

Commands:
  list    List material site reading material types
  show    Show material site reading material type details`,
	Example: `  # List material site reading material types
  xbe view material-site-reading-material-types list

  # Filter by material site
  xbe view material-site-reading-material-types list --material-site 123

  # Show a specific mapping
  xbe view material-site-reading-material-types show 456`,
}

func init() {
	viewCmd.AddCommand(materialSiteReadingMaterialTypesCmd)
}
