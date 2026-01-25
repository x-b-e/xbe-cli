package cli

import "github.com/spf13/cobra"

var applicationSettingsCmd = &cobra.Command{
	Use:     "application-settings",
	Aliases: []string{"application-setting"},
	Short:   "Browse and view application settings",
	Long: `Browse and view application settings.

Application settings are global key/value pairs that control platform behavior.
Access is restricted to admin users.

Commands:
  list    List application settings
  show    View full details of a specific setting`,
}

func init() {
	viewCmd.AddCommand(applicationSettingsCmd)
}
