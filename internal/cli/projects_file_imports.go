package cli

import "github.com/spf13/cobra"

var projectsFileImportsCmd = &cobra.Command{
	Use:     "projects-file-imports",
	Aliases: []string{"projects-file-import"},
	Short:   "Browse projects file imports",
	Long: `Browse projects file imports.

Projects file imports process project data from uploaded files. Imports run
asynchronously when created and capture status and results.`,
	Example: `  # List projects file imports
  xbe view projects-file-imports list

  # Show a projects file import
  xbe view projects-file-imports show 123`,
}

func init() {
	viewCmd.AddCommand(projectsFileImportsCmd)
}
