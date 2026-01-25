package cli

import "github.com/spf13/cobra"

func newMaterialSuppliersShowCmd() *cobra.Command {
	return newGenericShowCmd("material-suppliers")
}

func init() {
	materialSuppliersCmd.AddCommand(newMaterialSuppliersShowCmd())
}
