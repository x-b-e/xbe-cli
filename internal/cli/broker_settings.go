package cli

import "github.com/spf13/cobra"

var brokerSettingsCmd = &cobra.Command{
	Use:     "broker-settings",
	Aliases: []string{"broker-setting"},
	Short:   "View broker settings",
	Long:    "Commands for viewing broker settings.",
}

func init() {
	viewCmd.AddCommand(brokerSettingsCmd)
}
