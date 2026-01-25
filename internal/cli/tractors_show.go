package cli

import "github.com/spf13/cobra"

func newTractorsShowCmd() *cobra.Command {
	return newGenericShowCmd("tractors")
}

func init() {
	tractorsCmd.AddCommand(newTractorsShowCmd())
}
