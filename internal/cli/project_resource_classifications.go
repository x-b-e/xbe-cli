package cli

import "github.com/spf13/cobra"

var projectResourceClassificationsCmd = &cobra.Command{
	Use:   "project-resource-classifications",
	Short: "View project resource classifications",
	Long: `View project resource classifications on the XBE platform.

Project resource classifications define categories for project resources.
They are broker-scoped and can have parent-child relationships.

Commands:
  list    List project resource classifications`,
	Example: `  # List project resource classifications
  xbe view project-resource-classifications list

  # Filter by broker
  xbe view project-resource-classifications list --broker 123

  # Output as JSON
  xbe view project-resource-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(projectResourceClassificationsCmd)
}
