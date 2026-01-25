package cli

import "github.com/spf13/cobra"

var materialSiteMixingLotsCmd = &cobra.Command{
	Use:     "material-site-mixing-lots",
	Aliases: []string{"material-site-mixing-lot"},
	Short:   "View material site mixing lots",
	Long: `View material site mixing lots on the XBE platform.

Material site mixing lots represent contiguous mixing runs at a material site,
including production averages and the related material type.

Commands:
  list    List material site mixing lots with filtering
  show    Show material site mixing lot details`,
	Example: `  # List mixing lots
  xbe view material-site-mixing-lots list

  # Filter by material site
  xbe view material-site-mixing-lots list --material-site 123

  # Show a mixing lot
  xbe view material-site-mixing-lots show 456`,
}

func init() {
	viewCmd.AddCommand(materialSiteMixingLotsCmd)
}
