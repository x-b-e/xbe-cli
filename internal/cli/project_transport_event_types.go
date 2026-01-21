package cli

import "github.com/spf13/cobra"

var projectTransportEventTypesCmd = &cobra.Command{
	Use:   "project-transport-event-types",
	Short: "View project transport event types",
	Long: `View project transport event types.

Project transport event types define the types of events that can occur at transport stops.

Commands:
  list  List project transport event types`,
}

func init() {
	viewCmd.AddCommand(projectTransportEventTypesCmd)
}
