package cli

import "github.com/spf13/cobra"

var customerSettingsCmd = &cobra.Command{
	Use:     "customer-settings",
	Aliases: []string{"customer-setting"},
	Short:   "View customer settings",
	Long:    "Commands for viewing customer settings.",
}

func init() {
	viewCmd.AddCommand(customerSettingsCmd)
}
