package cli

import "github.com/spf13/cobra"

var ticketReportDispatchesCmd = &cobra.Command{
	Use:     "ticket-report-dispatches",
	Aliases: []string{"ticket-report-dispatch"},
	Short:   "Browse ticket report dispatches",
	Long:    "Commands for browsing ticket report dispatches.",
}

func init() {
	viewCmd.AddCommand(ticketReportDispatchesCmd)
}
