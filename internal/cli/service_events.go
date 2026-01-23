package cli

import "github.com/spf13/cobra"

var serviceEventsCmd = &cobra.Command{
	Use:     "service-events",
	Aliases: []string{"service-event"},
	Short:   "View service events",
	Long: `View service events on the XBE platform.

Service events capture shift check-in milestones (ready-to-work, work-start-at)
for tender job schedule shifts.

Commands:
  list    List service events
  show    Show service event details`,
	Example: `  # List service events
  xbe view service-events list

  # Show a service event
  xbe view service-events show 123`,
}

func init() {
	viewCmd.AddCommand(serviceEventsCmd)
}
