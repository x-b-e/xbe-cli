package cli

import "github.com/spf13/cobra"

func newExternalIdentificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("external-identifications")
}

func init() {
	externalIdentificationsCmd.AddCommand(newExternalIdentificationsShowCmd())
}
