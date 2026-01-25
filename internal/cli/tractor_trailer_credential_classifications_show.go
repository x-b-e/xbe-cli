package cli

import "github.com/spf13/cobra"

func newTractorTrailerCredentialClassificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("tractor-trailer-credential-classifications")
}

func init() {
	tractorTrailerCredentialClassificationsCmd.AddCommand(newTractorTrailerCredentialClassificationsShowCmd())
}
