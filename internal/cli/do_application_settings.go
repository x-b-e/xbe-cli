package cli

import "github.com/spf13/cobra"

var doApplicationSettingsCmd = &cobra.Command{
	Use:     "application-settings",
	Aliases: []string{"application-setting"},
	Short:   "Manage application settings",
	Long: `Create, update, and delete application settings.

Application settings are global key/value pairs that control platform behavior.
Access is restricted to admin users.`,
}

func init() {
	doCmd.AddCommand(doApplicationSettingsCmd)
}
