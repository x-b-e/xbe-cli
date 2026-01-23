package cli

import "github.com/spf13/cobra"

var doBrokerSettingsCmd = &cobra.Command{
	Use:     "broker-settings",
	Aliases: []string{"broker-setting"},
	Short:   "Manage broker settings",
	Long:    "Commands for updating broker settings.",
}

func init() {
	doCmd.AddCommand(doBrokerSettingsCmd)
}
