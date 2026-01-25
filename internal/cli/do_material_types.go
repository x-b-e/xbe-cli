package cli

import "github.com/spf13/cobra"

var doMaterialTypesCmd = &cobra.Command{
	Use:     "material-types",
	Aliases: []string{"material-type"},
	Short:   "Manage material types",
	Long:    `Create material types.`,
}

func init() {
	doCmd.AddCommand(doMaterialTypesCmd)
}
