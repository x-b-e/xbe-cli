package cli

import "github.com/spf13/cobra"

var notificationSubscriptionsCmd = &cobra.Command{
	Use:     "notification-subscriptions",
	Aliases: []string{"notification-subscription"},
	Short:   "Browse notification subscriptions",
	Long: `Browse notification subscriptions.

Notification subscriptions define which users receive specific notification types
and how they are delivered.

Commands:
  list    List notification subscriptions with filtering and pagination
  show    Show full details of a notification subscription`,
	Example: `  # List notification subscriptions
  xbe view notification-subscriptions list

  # Filter by user
  xbe view notification-subscriptions list --user 123

  # Show notification subscription details
  xbe view notification-subscriptions show 456`,
}

func init() {
	viewCmd.AddCommand(notificationSubscriptionsCmd)
}
