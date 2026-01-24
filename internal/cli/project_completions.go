package cli

import "github.com/spf13/cobra"

var projectCompletionsCmd = &cobra.Command{
	Use:     "project-completions",
	Aliases: []string{"project-completion"},
	Short:   "Browse project completions",
	Long: `Browse project completions.

Project completions record when a project is transitioned to complete status.

Commands:
  list    List project completions
  show    Show project completion details`,
	Example: `  # List project completions
  xbe view project-completions list

  # Show a project completion
  xbe view project-completions show 123`,
}

func init() {
	viewCmd.AddCommand(projectCompletionsCmd)
}
