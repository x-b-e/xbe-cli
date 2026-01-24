package cli

import "github.com/spf13/cobra"

var projectSubmissionsCmd = &cobra.Command{
	Use:     "project-submissions",
	Aliases: []string{"project-submission"},
	Short:   "View project submissions",
	Long: `View project submissions.

Project submissions transition projects from editing or rejected to submitted.

Commands:
  list    List project submissions
  show    Show project submission details`,
	Example: `  # List project submissions
  xbe view project-submissions list

  # Show a project submission
  xbe view project-submissions show 123

  # Output JSON
  xbe view project-submissions list --json`,
}

func init() {
	viewCmd.AddCommand(projectSubmissionsCmd)
}
