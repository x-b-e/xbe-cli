package cli

import "github.com/spf13/cobra"

var doLehmanRobertsApexViewpointTicketExportsCmd = &cobra.Command{
	Use:     "lehman-roberts-apex-viewpoint-ticket-exports",
	Aliases: []string{"lehman-roberts-apex-viewpoint-ticket-export"},
	Short:   "Generate Lehman Roberts Apex Viewpoint ticket exports",
	Long: `Generate Lehman Roberts Apex Viewpoint ticket exports.

Exports generate Viewpoint-compatible ticket CSVs for the Lehman Roberts branch.

Commands:
  create    Create a ticket export`,
	Example: `  # Create a ticket export
  xbe do lehman-roberts-apex-viewpoint-ticket-exports create \
    --template-name lrJWSCash \
    --sale-date-min 2025-01-01 \
    --sale-date-max 2025-01-31`,
}

func init() {
	doCmd.AddCommand(doLehmanRobertsApexViewpointTicketExportsCmd)
}
