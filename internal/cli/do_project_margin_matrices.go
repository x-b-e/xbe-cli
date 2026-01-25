package cli

import "github.com/spf13/cobra"

var doProjectMarginMatricesCmd = &cobra.Command{
	Use:     "project-margin-matrices",
	Aliases: []string{"project-margin-matrix"},
	Short:   "Manage project margin matrices",
	Long: `Manage project margin matrices on the XBE platform.

Project margin matrices provide scenario-based margin data for a project.

Commands:
  create    Create a project margin matrix
  delete    Delete a project margin matrix`,
	Example: `  # Create a project margin matrix
  xbe do project-margin-matrices create --project 123`,
}

func init() {
	doCmd.AddCommand(doProjectMarginMatricesCmd)
}
