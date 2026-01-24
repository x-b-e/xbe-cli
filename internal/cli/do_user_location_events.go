package cli

import "github.com/spf13/cobra"

var doUserLocationEventsCmd = &cobra.Command{
	Use:     "user-location-events",
	Aliases: []string{"user-location-event"},
	Short:   "Manage user location events",
	Long: `Manage user location events.

Commands:
  create    Create a user location event
  update    Update a user location event
  delete    Delete a user location event`,
	Example: `  # Create a user location event
  xbe do user-location-events create --user 123 --provenance gps --event-latitude 40.0 --event-longitude -74.0`,
}

func init() {
	doCmd.AddCommand(doUserLocationEventsCmd)
}
