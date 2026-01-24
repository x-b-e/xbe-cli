package cli

import "github.com/spf13/cobra"

var projectTrailerClassificationsCmd = &cobra.Command{
	Use:     "project-trailer-classifications",
	Aliases: []string{"project-trailer-classification"},
	Short:   "View project trailer classifications",
	Long: `Browse project trailer classifications.

Project trailer classifications associate trailer classifications with projects.
They can optionally link to project labor classifications.

Commands:
  list    List project trailer classifications
  show    Show project trailer classification details`,
	Example: `  # List project trailer classifications
  xbe view project-trailer-classifications list

  # Show a project trailer classification
  xbe view project-trailer-classifications show 123`,
}

func init() {
	viewCmd.AddCommand(projectTrailerClassificationsCmd)
}
