package cli

import "github.com/spf13/cobra"

var projectCancellationsCmd = &cobra.Command{
	Use:     "project-cancellations",
	Aliases: []string{"project-cancellation"},
	Short:   "Browse project cancellations",
	Long: `Browse project cancellations.

Project cancellations record when a project is transitioned to cancelled status.

Commands:
  list    List project cancellations
  show    Show project cancellation details`,
	Example: `  # List project cancellations
  xbe view project-cancellations list

  # Show a project cancellation
  xbe view project-cancellations show 123`,
}

func init() {
	viewCmd.AddCommand(projectCancellationsCmd)
}
