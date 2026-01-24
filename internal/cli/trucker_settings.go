package cli

import "github.com/spf13/cobra"

var truckerSettingsCmd = &cobra.Command{
	Use:     "trucker-settings",
	Aliases: []string{"trucker-setting"},
	Short:   "View trucker settings",
	Long:    "Commands for viewing trucker settings.",
}

func init() {
	viewCmd.AddCommand(truckerSettingsCmd)
}
