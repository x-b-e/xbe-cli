package cli

import "github.com/spf13/cobra"

var doPlatformStatusesCmd = &cobra.Command{
	Use:   "platform-statuses",
	Short: "Manage platform status updates",
	Long:  `Create, update, and delete platform status updates.`,
}

func init() {
	doCmd.AddCommand(doPlatformStatusesCmd)
}
