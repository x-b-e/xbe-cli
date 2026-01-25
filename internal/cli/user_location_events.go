package cli

import "github.com/spf13/cobra"

var userLocationEventsCmd = &cobra.Command{
	Use:     "user-location-events",
	Aliases: []string{"user-location-event"},
	Short:   "View user location events",
	Long: `Commands for viewing user location events.

User location events capture user-reported latitude/longitude with a
provenance (gps/map). Use list to browse events and show for full details.`,
	Example: `  # List user location events
  xbe view user-location-events list

  # Filter by user
  xbe view user-location-events list --user 123

  # Show a user location event
  xbe view user-location-events show 456`,
}

func init() {
	viewCmd.AddCommand(userLocationEventsCmd)
}
