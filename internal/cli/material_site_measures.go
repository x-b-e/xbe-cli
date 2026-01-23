package cli

import "github.com/spf13/cobra"

var materialSiteMeasuresCmd = &cobra.Command{
	Use:     "material-site-measures",
	Aliases: []string{"material-site-measure"},
	Short:   "View material site measures",
	Long: `View material site measures on the XBE platform.

Material site measures define the measurement types used for
material site readings (e.g., temperature or tons per hour).

Commands:
  list    List material site measures with filtering
  show    Show material site measure details`,
	Example: `  # List material site measures
  xbe view material-site-measures list

  # Filter by slug
  xbe view material-site-measures list --slug "mixing-temperature"

  # Show a material site measure
  xbe view material-site-measures show 123`,
}

func init() {
	viewCmd.AddCommand(materialSiteMeasuresCmd)
}
