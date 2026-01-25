package cli

import "github.com/spf13/cobra"

var doUserDeviceLocationTrackingRequestsCmd = &cobra.Command{
	Use:   "user-device-location-tracking-requests",
	Short: "Request user device location tracking",
	Long: `Request user device location tracking.

User device location tracking requests send push notifications to a user's
devices to start or stop location tracking.

Commands:
  create    Send a location tracking request`,
	Example: `  # Request the default (normal) tracking start
  xbe do user-device-location-tracking-requests create --user 123

  # Request continuous tracking start
  xbe do user-device-location-tracking-requests create \
    --user 123 \
    --location-tracking-kind continuous \
    --location-tracking-action start

  # Request tracking stop
  xbe do user-device-location-tracking-requests create \
    --user 123 \
    --location-tracking-action stop`,
}

func init() {
	doCmd.AddCommand(doUserDeviceLocationTrackingRequestsCmd)
}
