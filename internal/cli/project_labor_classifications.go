package cli

import "github.com/spf13/cobra"

var projectLaborClassificationsCmd = &cobra.Command{
	Use:   "project-labor-classifications",
	Short: "View project labor classifications",
	Long: `View project labor classifications on the XBE platform.

Project labor classifications link projects to labor classifications and
capture hourly rates used for prevailing wage calculations.

Commands:
  list    List project labor classifications
  show    Show project labor classification details`,
	Example: `  # List project labor classifications
  xbe view project-labor-classifications list

  # Filter by project
  xbe view project-labor-classifications list --project 123

  # Filter by labor classification
  xbe view project-labor-classifications list --labor-classification 456

  # Show a project labor classification
  xbe view project-labor-classifications show 789`,
}

func init() {
	viewCmd.AddCommand(projectLaborClassificationsCmd)
}
