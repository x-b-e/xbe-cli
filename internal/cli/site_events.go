package cli

import "github.com/spf13/cobra"

var siteEventsCmd = &cobra.Command{
	Use:     "site-events",
	Aliases: []string{"site-event"},
	Short:   "View site events",
	Long: `Commands for viewing site events.

Site events record arrivals, work start/stop events, and departures at job,
material, or parking sites. Use list to browse events and show for full details.`,
	Example: `  # List site events
  xbe view site-events list

  # Filter by event type
  xbe view site-events list --event-type start_work

  # Show a site event
  xbe view site-events show 123`,
}

func init() {
	viewCmd.AddCommand(siteEventsCmd)
}
