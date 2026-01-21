package cli

import "github.com/spf13/cobra"

var doProjectTransportEventTypesCmd = &cobra.Command{
	Use:   "project-transport-event-types",
	Short: "Manage project transport event types",
	Long: `Create, update, and delete project transport event types.

Project transport event types define the types of events that can occur at transport stops.

Commands:
  create  Create a new project transport event type
  update  Update an existing project transport event type
  delete  Delete a project transport event type`,
	Example: `  # Create a project transport event type
  xbe do project-transport-event-types create --name "Pickup" --code "PU" --broker 123

  # Update a project transport event type
  xbe do project-transport-event-types update 456 --name "Updated Name"

  # Delete a project transport event type
  xbe do project-transport-event-types delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectTransportEventTypesCmd)
}
