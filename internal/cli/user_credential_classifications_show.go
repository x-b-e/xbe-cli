package cli

import "github.com/spf13/cobra"

func newUserCredentialClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("user-credential-classifications")
}

func init() {
	userCredentialClassificationsCmd.AddCommand(newUserCredentialClassificationsShowCmd())
}
