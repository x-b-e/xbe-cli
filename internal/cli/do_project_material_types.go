package cli

import "github.com/spf13/cobra"

var doProjectMaterialTypesCmd = &cobra.Command{
	Use:     "project-material-types",
	Aliases: []string{"project-material-type"},
	Short:   "Manage project material types",
	Long:    "Commands for creating, updating, and deleting project material types.",
}

func init() {
	doCmd.AddCommand(doProjectMaterialTypesCmd)
}
