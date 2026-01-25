package cli

import "github.com/spf13/cobra"

func newPublicPraisesShowCmd() *cobra.Command {
	return newGenericShowCmd("public-praises")
}

func init() {
	publicPraisesCmd.AddCommand(newPublicPraisesShowCmd())
}
