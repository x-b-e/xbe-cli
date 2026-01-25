package cli

import "github.com/spf13/cobra"

var doMaterialSiteUnavailabilitiesCmd = &cobra.Command{
	Use:     "material-site-unavailabilities",
	Aliases: []string{"material-site-unavailability"},
	Short:   "Manage material site unavailabilities",
	Long: `Create, update, and delete material site unavailabilities.

Material site unavailabilities track planned or unplanned downtime windows
for material sites.

Commands:
  create    Create a new material site unavailability
  update    Update an existing material site unavailability
  delete    Delete a material site unavailability`,
}

func init() {
	doCmd.AddCommand(doMaterialSiteUnavailabilitiesCmd)
}
