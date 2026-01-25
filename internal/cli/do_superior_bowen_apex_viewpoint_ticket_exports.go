package cli

import "github.com/spf13/cobra"

var doSuperiorBowenApexViewpointTicketExportsCmd = &cobra.Command{
	Use:     "superior-bowen-apex-viewpoint-ticket-exports",
	Aliases: []string{"superior-bowen-apex-viewpoint-ticket-export"},
	Short:   "Generate Superior Bowen Apex Viewpoint ticket exports",
	Long: `Generate Superior Bowen Apex Viewpoint ticket exports.

Exports generate Viewpoint-compatible ticket CSVs for the Superior Bowen branch.

Commands:
  create    Create a ticket export`,
	Example: `  # Create a ticket export
  xbe do superior-bowen-apex-viewpoint-ticket-exports create \
    --sale-date-min 2025-01-01 \
    --sale-date-max 2025-01-31`,
}

func init() {
	doCmd.AddCommand(doSuperiorBowenApexViewpointTicketExportsCmd)
}
