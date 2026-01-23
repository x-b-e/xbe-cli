package cli

import "github.com/spf13/cobra"

var doServiceEventsCmd = &cobra.Command{
	Use:   "service-events",
	Short: "Manage service events",
	Long: `Manage service events on the XBE platform.

Commands:
  create    Create a service event
  update    Update a service event
  delete    Delete a service event`,
	Example: `  # Create a service event
  xbe do service-events create \
    --tender-job-schedule-shift 123 \
    --occurred-at 2026-01-23T12:00:00Z \
    --kind ready_to_work

  # Update a service event note
  xbe do service-events update 123 --note "Arrived on site"

  # Delete a service event (requires --confirm)
  xbe do service-events delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doServiceEventsCmd)
}
