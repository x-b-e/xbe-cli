package cli

import "github.com/spf13/cobra"

var digitalFleetTicketEventsCmd = &cobra.Command{
	Use:     "digital-fleet-ticket-events",
	Aliases: []string{"digital-fleet-ticket-event"},
	Short:   "Browse digital fleet ticket events",
	Long: `Browse digital fleet ticket events.

Digital fleet ticket events capture telematics ticket events ingested from
Digital Fleet.

Commands:
  list    List digital fleet ticket events with filtering and pagination
  show    Show full details of a digital fleet ticket event`,
	Example: `  # List digital fleet ticket events
  xbe view digital-fleet-ticket-events list

  # Show digital fleet ticket event details
  xbe view digital-fleet-ticket-events show 123`,
}

func init() {
	viewCmd.AddCommand(digitalFleetTicketEventsCmd)
}
