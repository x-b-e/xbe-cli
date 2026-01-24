package cli

import "github.com/spf13/cobra"

var projectMarginMatricesCmd = &cobra.Command{
	Use:     "project-margin-matrices",
	Aliases: []string{"project-margin-matrix"},
	Short:   "Browse project margin matrices",
	Long: `Browse project margin matrices.

Project margin matrices summarize scenario-based margin data for a project.

Commands:
  list    List project margin matrices
  show    Show project margin matrix details`,
	Example: `  # List project margin matrices
  xbe view project-margin-matrices list

  # Show a project margin matrix
  xbe view project-margin-matrices show 123`,
}

func init() {
	viewCmd.AddCommand(projectMarginMatricesCmd)
}
