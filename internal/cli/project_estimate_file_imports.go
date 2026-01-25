package cli

import "github.com/spf13/cobra"

var projectEstimateFileImportsCmd = &cobra.Command{
	Use:     "project-estimate-file-imports",
	Aliases: []string{"project-estimate-file-import"},
	Short:   "Browse project estimate file imports",
	Long: `Browse project estimate file imports.

Project estimate file imports process estimate data from uploaded files for a
project. Imports run asynchronously when created and are not persisted.`,
	Example: `  # List project estimate file imports (typically empty)
  xbe view project-estimate-file-imports list`,
}

func init() {
	viewCmd.AddCommand(projectEstimateFileImportsCmd)
}
