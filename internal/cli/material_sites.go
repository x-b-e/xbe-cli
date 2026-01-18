package cli

import "github.com/spf13/cobra"

var materialSitesCmd = &cobra.Command{
	Use:   "material-sites",
	Short: "View material sites",
	Long: `View material sites on the XBE platform.

Material sites are source locations for materials - plants, quarries,
or stockpile locations where trucks pick up materials.

Commands:
  list    List material sites with filtering`,
	Example: `  # List material sites
  xbe view material-sites list

  # Search by name
  xbe view material-sites list --name "Plant"

  # List active material sites only
  xbe view material-sites list --active`,
}

func init() {
	viewCmd.AddCommand(materialSitesCmd)
}
