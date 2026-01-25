package cli

import "github.com/spf13/cobra"

var doSiteEventsCmd = &cobra.Command{
	Use:   "site-events",
	Short: "Manage site events",
	Long: `Manage site events on the XBE platform.

Commands:
  create   Create a new site event
  update   Update an existing site event
  delete   Delete a site event (requires --confirm)`,
	Example: `  # Create a site event
  xbe do site-events create \
    --event-type start_work \
    --event-kind load \
    --event-at 2025-01-01T12:00:00Z \
    --event-latitude 41.8781 \
    --event-longitude -87.6298 \
    --material-transaction 123 \
    --event-site-type material-sites \
    --event-site-id 456

  # Update a site event
  xbe do site-events update 123 --event-details "Updated details"

  # Delete a site event
  xbe do site-events delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doSiteEventsCmd)
}
