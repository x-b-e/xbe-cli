package cli

import "github.com/spf13/cobra"

func newMaterialTypesShowCmd() *cobra.Command {
	return newGenericShowCmd("material-types")
}

func init() {
	materialTypesCmd.AddCommand(newMaterialTypesShowCmd())
}
