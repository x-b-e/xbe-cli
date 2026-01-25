package cli

import "github.com/spf13/cobra"

var notificationsCmd = &cobra.Command{
	Use:     "notifications",
	Aliases: []string{"notification"},
	Short:   "Browse notifications",
	Long: `Browse notifications for the current user.

Notifications capture delivery status and payload details across the platform.

Commands:
  list    List notifications
  show    Show notification details`,
	Example: `  # List notifications
  xbe view notifications list

  # Show a notification
  xbe view notifications show 123

  # Output as JSON
  xbe view notifications list --json`,
}

func init() {
	viewCmd.AddCommand(notificationsCmd)
}
