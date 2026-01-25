package cli

import "github.com/spf13/cobra"

var superiorBowenApexViewpointTicketExportsCmd = &cobra.Command{
	Use:     "superior-bowen-apex-viewpoint-ticket-exports",
	Aliases: []string{"superior-bowen-apex-viewpoint-ticket-export"},
	Short:   "Browse Superior Bowen Apex Viewpoint ticket exports",
	Long: `Browse Superior Bowen Apex Viewpoint ticket exports.

These exports generate Viewpoint-compatible ticket CSVs for the Superior Bowen
branch.

Commands:
  list    List Superior Bowen Apex Viewpoint ticket exports`,
	Example: `  # List exports
  xbe view superior-bowen-apex-viewpoint-ticket-exports list`,
}

func init() {
	viewCmd.AddCommand(superiorBowenApexViewpointTicketExportsCmd)
}
