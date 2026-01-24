package cli

import "github.com/spf13/cobra"

var doTicketReportImportsCmd = &cobra.Command{
	Use:     "ticket-report-imports",
	Aliases: []string{"ticket-report-import"},
	Short:   "Manage ticket report imports",
	Long:    "Commands for creating and deleting ticket report imports.",
}

func init() {
	doCmd.AddCommand(doTicketReportImportsCmd)
}
