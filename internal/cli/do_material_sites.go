package cli

import "github.com/spf13/cobra"

var doMaterialSitesCmd = &cobra.Command{
	Use:   "material-sites",
	Short: "Manage material sites",
	Long: `Create, update, and delete material sites.

Material sites are source locations for materials - plants, quarries,
or stockpile locations where trucks pick up materials.

Commands:
  create    Create a new material site
  update    Update an existing material site
  delete    Delete a material site`,
}

func init() {
	doCmd.AddCommand(doMaterialSitesCmd)
}
