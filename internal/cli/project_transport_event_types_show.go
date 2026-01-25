package cli

import "github.com/spf13/cobra"

func newProjectTransportEventTypesShowCmd() *cobra.Command {
	return newGenericShowCmd("project-transport-event-types")
}

func init() {
	projectTransportEventTypesCmd.AddCommand(newProjectTransportEventTypesShowCmd())
}
