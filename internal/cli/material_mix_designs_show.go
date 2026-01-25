package cli

import "github.com/spf13/cobra"

func newMaterialMixDesignsShowCmd() *cobra.Command {
	return newGenericShowCmd("material-mix-designs")
}

func init() {
	materialMixDesignsCmd.AddCommand(newMaterialMixDesignsShowCmd())
}
