package cli

import "github.com/spf13/cobra"

var placesCmd = &cobra.Command{
	Use:     "places",
	Aliases: []string{"place"},
	Short:   "Lookup place details",
	Long: `Lookup place details by Google Place ID.

Places return the formatted address and coordinates for a specific
Google Place ID.

Commands:
  show    Show place details`,
	Example: `  # Show a place
  xbe view places show ChIJD7fiBh9u5kcRYJSMaMOCCwQ

  # Get JSON output
  xbe view places show ChIJD7fiBh9u5kcRYJSMaMOCCwQ --json`,
}

func init() {
	viewCmd.AddCommand(placesCmd)
}
