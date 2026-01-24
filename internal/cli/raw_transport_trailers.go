package cli

import "github.com/spf13/cobra"

var rawTransportTrailersCmd = &cobra.Command{
	Use:     "raw-transport-trailers",
	Aliases: []string{"raw-transport-trailer"},
	Short:   "Browse raw transport trailers",
	Long: `Browse raw transport trailers on the XBE platform.

Raw transport trailers represent upstream trailer records imported from
transport systems for validation and processing.

Commands:
  list    List raw transport trailers
  show    Show raw transport trailer details`,
	Example: `  # List raw transport trailers
  xbe view raw-transport-trailers list

  # Show raw transport trailer details
  xbe view raw-transport-trailers show 123`,
}

func init() {
	viewCmd.AddCommand(rawTransportTrailersCmd)
}
