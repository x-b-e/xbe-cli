package cli

import "github.com/spf13/cobra"

var newslettersCmd = &cobra.Command{
	Use:   "newsletters",
	Short: "Browse and view newsletters",
	Long: `Browse and view published newsletters.

Newsletters contain market analysis, insights, and updates published by
brokers on the XBE platform. You can list newsletters with various filters
or view the full content of a specific newsletter.

Commands:
  list    List newsletters with filtering and pagination
  show    View the full content of a specific newsletter

Filtering:
  The list command supports extensive filtering options including:
  - Publication date ranges
  - Organization/broker
  - Public/private visibility
  - Text search`,
	Example: `  # List recent published newsletters
  xbe view newsletters list

  # Search newsletters by keyword
  xbe view newsletters list --q "market analysis"

  # Filter by broker
  xbe view newsletters list --broker-id 123

  # Filter by date range
  xbe view newsletters list --published-on-min 2024-01-01 --published-on-max 2024-12-31

  # Get results as JSON for scripting
  xbe view newsletters list --json --limit 10

  # View a specific newsletter
  xbe view newsletters show 456`,
}

func init() {
	viewCmd.AddCommand(newslettersCmd)
}
