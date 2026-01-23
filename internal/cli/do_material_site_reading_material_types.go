package cli

import "github.com/spf13/cobra"

var doMaterialSiteReadingMaterialTypesCmd = &cobra.Command{
	Use:   "material-site-reading-material-types",
	Short: "Manage material site reading material types",
	Long: `Create, update, and delete material site reading material types.

Material site reading material types map external material identifiers from
plant or system readings to internal material types for a specific material site.

Commands:
  create    Create a new material site reading material type
  update    Update an existing material site reading material type
  delete    Delete a material site reading material type`,
}

func init() {
	doCmd.AddCommand(doMaterialSiteReadingMaterialTypesCmd)
}
