package cli

import "github.com/spf13/cobra"

var doTicketReportDispatchesCmd = &cobra.Command{
	Use:     "ticket-report-dispatches",
	Aliases: []string{"ticket-report-dispatch"},
	Short:   "Manage ticket report dispatches",
	Long:    "Commands for creating and deleting ticket report dispatches.",
}

func init() {
	doCmd.AddCommand(doTicketReportDispatchesCmd)
}
