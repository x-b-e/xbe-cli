package cli

import "github.com/spf13/cobra"

var doTicketReportsCmd = &cobra.Command{
	Use:   "ticket-reports",
	Short: "Manage ticket reports",
	Long: `Create, update, and delete ticket reports.

Ticket reports track ticket files uploaded for dispatch, import, or validation.

Commands:
  create    Create a new ticket report
  update    Update an existing ticket report
  delete    Delete a ticket report`,
}

func init() {
	doCmd.AddCommand(doTicketReportsCmd)
}
