package cli

import "github.com/spf13/cobra"

func newTrailerCredentialsShowCmd() *cobra.Command {
	return newGenericShowCmd("trailer-credentials")
}

func init() {
	trailerCredentialsCmd.AddCommand(newTrailerCredentialsShowCmd())
}
