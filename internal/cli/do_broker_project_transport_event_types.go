package cli

import "github.com/spf13/cobra"

var doBrokerProjectTransportEventTypesCmd = &cobra.Command{
	Use:     "broker-project-transport-event-types",
	Aliases: []string{"broker-project-transport-event-type"},
	Short:   "Manage broker project transport event types",
	Long: `Create, update, and delete broker project transport event types.

Commands:
  create    Create a broker project transport event type
  update    Update a broker project transport event type
  delete    Delete a broker project transport event type`,
	Example: `  # Create a broker project transport event type
  xbe do broker-project-transport-event-types create --broker 123 --project-transport-event-type 456 --code "PICK"

  # Update a broker project transport event type
  xbe do broker-project-transport-event-types update 789 --code "DROP"

  # Delete a broker project transport event type (requires --confirm)
  xbe do broker-project-transport-event-types delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doBrokerProjectTransportEventTypesCmd)
}
