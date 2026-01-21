package cli

import "github.com/spf13/cobra"

var trailerClassificationsCmd = &cobra.Command{
	Use:   "trailer-classifications",
	Short: "View trailer classifications",
	Long: `View trailer classifications on the XBE platform.

Trailer classifications define types of trailers (e.g., end dump, belly dump,
flatbed) with their specifications like axle count and capacity. These are
used to match trailers to job requirements.

Commands:
  list    List trailer classifications`,
	Example: `  # List trailer classifications
  xbe view trailer-classifications list

  # Output as JSON
  xbe view trailer-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(trailerClassificationsCmd)
}
