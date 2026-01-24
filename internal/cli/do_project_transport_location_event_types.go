package cli

import "github.com/spf13/cobra"

var doProjectTransportLocationEventTypesCmd = &cobra.Command{
	Use:     "project-transport-location-event-types",
	Aliases: []string{"project-transport-location-event-type"},
	Short:   "Manage project transport location event types",
	Long: `Create and delete project transport location event types.

Commands:
  create    Create a project transport location event type
  delete    Delete a project transport location event type`,
}

func init() {
	doCmd.AddCommand(doProjectTransportLocationEventTypesCmd)
}
