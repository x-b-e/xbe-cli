package cli

import "github.com/spf13/cobra"

var doNotificationsCmd = &cobra.Command{
	Use:     "notifications",
	Aliases: []string{"notification"},
	Short:   "Manage notifications",
	Long:    "Commands for updating notifications.",
}

func init() {
	doCmd.AddCommand(doNotificationsCmd)
}
