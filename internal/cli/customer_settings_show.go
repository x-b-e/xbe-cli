package cli

import "github.com/spf13/cobra"

func newCustomerSettingsShowCmd() *cobra.Command {
	return newGenericShowCmd("customer-settings")
}

func init() {
	customerSettingsCmd.AddCommand(newCustomerSettingsShowCmd())
}
