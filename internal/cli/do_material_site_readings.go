package cli

import "github.com/spf13/cobra"

var doMaterialSiteReadingsCmd = &cobra.Command{
	Use:   "material-site-readings",
	Short: "Manage material site readings",
	Long: `Create, update, and delete material site readings.

Material site readings capture measured values for material site measures and
are used to track inventory and mixing data.

Commands:
  create    Create a new material site reading
  update    Update an existing material site reading
  delete    Delete a material site reading`,
}

func init() {
	doCmd.AddCommand(doMaterialSiteReadingsCmd)
}
