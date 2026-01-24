package cli

import "github.com/spf13/cobra"

var userLocationEstimatesCmd = &cobra.Command{
	Use:     "user-location-estimates",
	Aliases: []string{"user-location-estimate"},
	Short:   "Browse user location estimates",
	Long: `Browse user location estimates.

User location estimates return the most recent known location for a user based
on device, vehicle, and activity events.

Commands:
  list    List user location estimates (requires --user)`,
	Example: `  # Estimate a user's location
  xbe view user-location-estimates list --user 123

  # Estimate with a custom as-of time
  xbe view user-location-estimates list --user 123 --as-of 2025-01-01T12:00:00Z

  # Constrain the event window
  xbe view user-location-estimates list --user 123 \
    --earliest-event-at 2025-01-01T00:00:00Z \
    --latest-event-at 2025-01-02T00:00:00Z`,
}

func init() {
	viewCmd.AddCommand(userLocationEstimatesCmd)
}
