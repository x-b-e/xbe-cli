package cli

import "github.com/spf13/cobra"

var projectBidLocationsCmd = &cobra.Command{
	Use:     "project-bid-locations",
	Aliases: []string{"project-bid-location"},
	Short:   "View project bid locations",
	Long: `View project bid locations.

Project bid locations define geometry-based bid locations associated with
projects and their prediction subjects.

Commands:
  list    List project bid locations with filtering
  show    Show project bid location details`,
	Example: `  # List project bid locations
  xbe view project-bid-locations list

  # Show a project bid location
  xbe view project-bid-locations show 123`,
}

func init() {
	viewCmd.AddCommand(projectBidLocationsCmd)
}
