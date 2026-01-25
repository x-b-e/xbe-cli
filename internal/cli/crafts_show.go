package cli

import "github.com/spf13/cobra"

func newCraftsShowCmd() *cobra.Command {
	return newGenericShowCmd("crafts")
}

func init() {
	craftsCmd.AddCommand(newCraftsShowCmd())
}
