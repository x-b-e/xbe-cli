package cli

import "github.com/spf13/cobra"

var ticketReportsCmd = &cobra.Command{
	Use:   "ticket-reports",
	Short: "Browse ticket reports",
	Long: `Browse ticket reports on the XBE platform.

Ticket reports track ticket files uploaded for dispatch, import, or validation.

Commands:
  list    List ticket reports with filtering
  show    Show ticket report details`,
	Example: `  # List ticket reports
  xbe view ticket-reports list

  # Show a ticket report
  xbe view ticket-reports show 123`,
}

func init() {
	viewCmd.AddCommand(ticketReportsCmd)
}
