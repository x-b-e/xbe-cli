package cli

import "github.com/spf13/cobra"

func newCraftClassesShowCmd() *cobra.Command {
	return newGenericShowCmd("craft-classes")
}

func init() {
	craftClassesCmd.AddCommand(newCraftClassesShowCmd())
}
