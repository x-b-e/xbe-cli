package cli

import "github.com/spf13/cobra"

var projectTransportLocationEventTypesCmd = &cobra.Command{
	Use:     "project-transport-location-event-types",
	Aliases: []string{"project-transport-location-event-type"},
	Short:   "View project transport location event types",
	Long: `View project transport location event types.

Project transport location event types link transport locations to the event types
that can occur at those locations.

Commands:
  list    List project transport location event types
  show    Show project transport location event type details`,
	Example: `  # List location event types
  xbe view project-transport-location-event-types list

  # Show a specific location event type
  xbe view project-transport-location-event-types show 123

  # Output JSON
  xbe view project-transport-location-event-types list --json`,
}

func init() {
	viewCmd.AddCommand(projectTransportLocationEventTypesCmd)
}
