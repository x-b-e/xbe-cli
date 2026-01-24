package cli

import "github.com/spf13/cobra"

var versionEventsCmd = &cobra.Command{
	Use:     "version-events",
	Aliases: []string{"version-event"},
	Short:   "Browse version events",
	Long: `Browse version events.

Version events record change events that are exported to downstream integrations.

Commands:
  list    List version events with filtering and pagination
  show    Show full details of a version event`,
	Example: `  # List version events
  xbe view version-events list

  # Show version event details
  xbe view version-events show 123`,
}

func init() {
	viewCmd.AddCommand(versionEventsCmd)
}
