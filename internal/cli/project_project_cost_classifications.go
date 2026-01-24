package cli

import "github.com/spf13/cobra"

var projectProjectCostClassificationsCmd = &cobra.Command{
	Use:   "project-project-cost-classifications",
	Short: "View project project cost classifications",
	Long: `View project project cost classifications on the XBE platform.

Project project cost classifications link a project to a project cost classification
and optionally override the classification name for that project.

Commands:
  list    List project project cost classifications
  show    Show project project cost classification details`,
	Example: `  # List project project cost classifications
  xbe view project-project-cost-classifications list

  # Filter by project
  xbe view project-project-cost-classifications list --project 123

  # Show details
  xbe view project-project-cost-classifications show 456

  # Output as JSON
  xbe view project-project-cost-classifications list --json`,
}

func init() {
	viewCmd.AddCommand(projectProjectCostClassificationsCmd)
}
