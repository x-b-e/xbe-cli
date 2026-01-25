package cli

import "github.com/spf13/cobra"

func newUserCredentialsShowCmd() *cobra.Command {
	return newGenericShowCmd("user-credentials")
}

func init() {
	userCredentialsCmd.AddCommand(newUserCredentialsShowCmd())
}
