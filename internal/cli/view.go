package cli

import "github.com/spf13/cobra"

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Browse and view XBE content",
	Long: `Browse and view XBE content.

The view command provides read-only access to XBE platform data including
newsletters and broker information. All view commands support:

  --json       Output in JSON format for programmatic use
  --no-auth    Access public content without authentication
  --limit      Control the number of results returned
  --offset     Paginate through large result sets

Content Types:
  newsletters  Published market newsletters with analysis and insights
  brokers      Broker/branch information and metadata`,
	Example: `  # Browse newsletters
  xbe view newsletters list
  xbe view newsletters show 123

  # Browse brokers
  xbe view brokers list`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
