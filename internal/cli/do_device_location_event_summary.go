package cli

import "github.com/spf13/cobra"

var doDeviceLocationEventSummaryCmd = &cobra.Command{
	Use:   "device-location-event-summary",
	Short: "Generate device location event summaries",
	Long: `Generate device location event summaries on the XBE platform.

Commands:
  create    Generate a device location event summary`,
	Example: `  # Generate a device location event summary grouped by device
  xbe summarize device-location-event-summary create --group-by device --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31

  # Summary by user
  xbe summarize device-location-event-summary create --group-by user --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31

  # Summary by device platform
  xbe summarize device-location-event-summary create --group-by device_platform --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31`,
}

func init() {
	summarizeCmd.AddCommand(doDeviceLocationEventSummaryCmd)
}
