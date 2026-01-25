package cli

import "github.com/spf13/cobra"

var lineupSummaryRequestsCmd = &cobra.Command{
	Use:     "lineup-summary-requests",
	Aliases: []string{"lineup-summary-request"},
	Short:   "Browse lineup summary requests",
	Long: `Browse lineup summary requests.

Lineup summary requests trigger lineup summary emails for a broker or customer.

Commands:
  list    List lineup summary requests
  show    Show lineup summary request details`,
	Example: `  # List lineup summary requests
  xbe view lineup-summary-requests list

  # Show a lineup summary request
  xbe view lineup-summary-requests show 123`,
}

func init() {
	viewCmd.AddCommand(lineupSummaryRequestsCmd)
}
