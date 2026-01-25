package cli

import "github.com/spf13/cobra"

func newTractorCredentialsShowCmd() *cobra.Command {
	return newGenericShowCmd("tractor-credentials")
}

func init() {
	tractorCredentialsCmd.AddCommand(newTractorCredentialsShowCmd())
}
