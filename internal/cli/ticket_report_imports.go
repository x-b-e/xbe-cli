package cli

import "github.com/spf13/cobra"

var ticketReportImportsCmd = &cobra.Command{
	Use:     "ticket-report-imports",
	Aliases: []string{"ticket-report-import"},
	Short:   "Browse ticket report imports",
	Long:    "Commands for browsing ticket report imports.",
}

func init() {
	viewCmd.AddCommand(ticketReportImportsCmd)
}
