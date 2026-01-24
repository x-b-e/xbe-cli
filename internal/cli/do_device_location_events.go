package cli

import "github.com/spf13/cobra"

var doDeviceLocationEventsCmd = &cobra.Command{
	Use:     "device-location-events",
	Aliases: []string{"device-location-event"},
	Short:   "Record device location events",
	Long: `Record device location events reported by devices.

Commands:
  create    Create a device location event`,
	Example: `  # Create a device location event
  xbe do device-location-events create --device-identifier "ios:ABC123" --payload '{"uuid":"evt-1","timestamp":"2025-01-01T00:00:00Z","activity":{"type":"walking"},"coords":{"latitude":40.0,"longitude":-74.0}}'`,
}

func init() {
	doCmd.AddCommand(doDeviceLocationEventsCmd)
}
