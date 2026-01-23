package cli

import "github.com/spf13/cobra"

var materialTypeUnavailabilitiesCmd = &cobra.Command{
	Use:     "material-type-unavailabilities",
	Aliases: []string{"material-type-unavailability"},
	Short:   "View material type unavailabilities",
	Long: `View material type unavailabilities on the XBE platform.

Material type unavailabilities define time windows when a supplier-specific
material type cannot be used for planning or scheduling.

Commands:
  list    List material type unavailabilities
  show    Show material type unavailability details`,
	Example: `  # List material type unavailabilities
  xbe view material-type-unavailabilities list

  # Show one
  xbe view material-type-unavailabilities show 123`,
}

func init() {
	viewCmd.AddCommand(materialTypeUnavailabilitiesCmd)
}
