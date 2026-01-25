package cli

import "github.com/spf13/cobra"

func newExternalIdentificationTypesShowCmd() *cobra.Command {
	return newGenericShowCmd("external-identification-types")
}

func init() {
	externalIdentificationTypesCmd.AddCommand(newExternalIdentificationTypesShowCmd())
}
