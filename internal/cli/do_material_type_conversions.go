package cli

import "github.com/spf13/cobra"

var doMaterialTypeConversionsCmd = &cobra.Command{
	Use:     "material-type-conversions",
	Aliases: []string{"material-type-conversion"},
	Short:   "Manage material type conversions",
	Long:    "Commands for creating, updating, and deleting material type conversions.",
}

func init() {
	doCmd.AddCommand(doMaterialTypeConversionsCmd)
}
