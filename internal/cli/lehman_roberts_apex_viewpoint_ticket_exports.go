package cli

import "github.com/spf13/cobra"

var lehmanRobertsApexViewpointTicketExportsCmd = &cobra.Command{
	Use:     "lehman-roberts-apex-viewpoint-ticket-exports",
	Aliases: []string{"lehman-roberts-apex-viewpoint-ticket-export"},
	Short:   "Browse Lehman Roberts Apex Viewpoint ticket exports",
	Long: `Browse Lehman Roberts Apex Viewpoint ticket exports.

These exports generate Viewpoint-compatible ticket CSVs for the Lehman Roberts
branch.

Commands:
  list    List Lehman Roberts Apex Viewpoint ticket exports`,
	Example: `  # List exports
  xbe view lehman-roberts-apex-viewpoint-ticket-exports list`,
}

func init() {
	viewCmd.AddCommand(lehmanRobertsApexViewpointTicketExportsCmd)
}
