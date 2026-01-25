package cli

import "github.com/spf13/cobra"

func newCertificationRequirementsShowCmd() *cobra.Command {
	return newGenericShowCmd("certification-requirements")
}

func init() {
	certificationRequirementsCmd.AddCommand(newCertificationRequirementsShowCmd())
}
