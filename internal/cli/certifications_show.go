package cli

import "github.com/spf13/cobra"

func newCertificationsShowCmd() *cobra.Command {
	return newGenericShowCmd("certifications")
}

func init() {
	certificationsCmd.AddCommand(newCertificationsShowCmd())
}
