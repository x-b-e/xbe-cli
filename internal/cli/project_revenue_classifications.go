package cli

import "github.com/spf13/cobra"

var projectRevenueClassificationsCmd = &cobra.Command{
	Use:   "project-revenue-classifications",
	Short: "View project revenue classifications",
	Long: `View project revenue classifications on the XBE platform.

Project revenue classifications define the hierarchy of revenue categories for projects.
They are broker-scoped and can have parent-child relationships.

Commands:
  list    List project revenue classifications`,
	Example: `  # List project revenue classifications
  xbe view project-revenue-classifications list

  # Filter by broker
  xbe view project-revenue-classifications list --broker 123

  # Output as JSON
  xbe view project-revenue-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(projectRevenueClassificationsCmd)
}
