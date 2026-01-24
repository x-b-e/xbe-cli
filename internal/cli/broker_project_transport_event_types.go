package cli

import "github.com/spf13/cobra"

var brokerProjectTransportEventTypesCmd = &cobra.Command{
	Use:     "broker-project-transport-event-types",
	Aliases: []string{"broker-project-transport-event-type"},
	Short:   "View broker project transport event types",
	Long: `View broker project transport event types.

Broker project transport event types define broker-specific codes for transport
event types.

Commands:
  list    List broker project transport event types
  show    Show broker project transport event type details`,
	Example: `  # List broker project transport event types
  xbe view broker-project-transport-event-types list

  # Show a broker project transport event type
  xbe view broker-project-transport-event-types show 123

  # Output JSON
  xbe view broker-project-transport-event-types list --json`,
}

func init() {
	viewCmd.AddCommand(brokerProjectTransportEventTypesCmd)
}
