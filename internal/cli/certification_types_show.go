package cli

import "github.com/spf13/cobra"

func newCertificationTypesShowCmd() *cobra.Command {
	return newGenericShowCmd("certification-types")
}

func init() {
	certificationTypesCmd.AddCommand(newCertificationTypesShowCmd())
}
