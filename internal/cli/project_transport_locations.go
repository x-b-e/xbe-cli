package cli

import "github.com/spf13/cobra"

var projectTransportLocationsCmd = &cobra.Command{
	Use:   "project-transport-locations",
	Short: "View project transport locations",
	Long: `View project transport locations.

Project transport locations represent pickup, delivery, and staging locations
used in transport planning.

Commands:
  list  List project transport locations
  show  Show project transport location details`,
	Example: `  # List project transport locations
  xbe view project-transport-locations list

  # Show a project transport location
  xbe view project-transport-locations show 123`,
}

func init() {
	viewCmd.AddCommand(projectTransportLocationsCmd)
}
