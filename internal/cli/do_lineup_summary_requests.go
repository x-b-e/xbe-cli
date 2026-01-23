package cli

import "github.com/spf13/cobra"

var doLineupSummaryRequestsCmd = &cobra.Command{
	Use:     "lineup-summary-requests",
	Aliases: []string{"lineup-summary-request"},
	Short:   "Request lineup summaries",
	Long:    "Commands for requesting lineup summary emails.",
}

func init() {
	doCmd.AddCommand(doLineupSummaryRequestsCmd)
}
