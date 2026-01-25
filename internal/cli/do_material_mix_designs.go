package cli

import "github.com/spf13/cobra"

var doMaterialMixDesignsCmd = &cobra.Command{
	Use:     "material-mix-designs",
	Aliases: []string{"material-mix-design"},
	Short:   "Manage material mix designs",
	Long:    "Commands for creating, updating, and deleting material mix designs.",
}

func init() {
	doCmd.AddCommand(doMaterialMixDesignsCmd)
}
