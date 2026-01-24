package cli

import "github.com/spf13/cobra"

var doTruckerSettingsCmd = &cobra.Command{
	Use:     "trucker-settings",
	Aliases: []string{"trucker-setting"},
	Short:   "Manage trucker settings",
	Long:    "Commands for creating and updating trucker settings.",
}

func init() {
	doCmd.AddCommand(doTruckerSettingsCmd)
}
