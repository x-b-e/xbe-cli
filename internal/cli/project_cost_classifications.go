package cli

import "github.com/spf13/cobra"

var projectCostClassificationsCmd = &cobra.Command{
	Use:   "project-cost-classifications",
	Short: "View project cost classifications",
	Long: `View project cost classifications on the XBE platform.

Project cost classifications define the hierarchy of cost categories for projects.
They are broker-scoped and can have parent-child relationships.

Commands:
  list    List project cost classifications`,
	Example: `  # List project cost classifications
  xbe view project-cost-classifications list

  # Filter by broker
  xbe view project-cost-classifications list --broker 123

  # Output as JSON
  xbe view project-cost-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(projectCostClassificationsCmd)
}
