package cli

import "github.com/spf13/cobra"

var doProjectCancellationsCmd = &cobra.Command{
	Use:   "project-cancellations",
	Short: "Cancel projects",
	Long: `Cancel projects on the XBE platform.

Cancellations transition projects to cancelled status. Only projects in approved
status can be cancelled.

Commands:
  create    Cancel a project`,
	Example: `  # Cancel a project
  xbe do project-cancellations create --project 123 --comment "Customer withdrew"`,
}

func init() {
	doCmd.AddCommand(doProjectCancellationsCmd)
}
