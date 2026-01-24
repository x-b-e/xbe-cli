package cli

import "github.com/spf13/cobra"

var doProjectCompletionsCmd = &cobra.Command{
	Use:   "project-completions",
	Short: "Complete projects",
	Long: `Complete projects on the XBE platform.

Completions transition projects to complete status. Only projects in approved
status can be completed.

Commands:
  create    Complete a project`,
	Example: `  # Complete a project
  xbe do project-completions create --project 123 --comment "Finalized"`,
}

func init() {
	doCmd.AddCommand(doProjectCompletionsCmd)
}
