package cli

import "github.com/spf13/cobra"

var doCustomerSettingsCmd = &cobra.Command{
	Use:     "customer-settings",
	Aliases: []string{"customer-setting"},
	Short:   "Manage customer settings",
	Long:    "Commands for updating customer settings.",
}

func init() {
	doCmd.AddCommand(doCustomerSettingsCmd)
}
