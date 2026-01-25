package cli

import "github.com/spf13/cobra"

var materialSiteUnavailabilitiesCmd = &cobra.Command{
	Use:     "material-site-unavailabilities",
	Aliases: []string{"material-site-unavailability"},
	Short:   "View material site unavailabilities",
	Long: `View material site unavailabilities.

Material site unavailabilities track planned or unplanned downtime windows
for material sites.

Commands:
  list    List material site unavailabilities with filtering
  show    Show material site unavailability details`,
	Example: `  # List material site unavailabilities
  xbe view material-site-unavailabilities list

  # Filter by material site
  xbe view material-site-unavailabilities list --material-site 123

  # Show an unavailability
  xbe view material-site-unavailabilities show 456`,
}

func init() {
	viewCmd.AddCommand(materialSiteUnavailabilitiesCmd)
}
