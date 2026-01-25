package cli

import "github.com/spf13/cobra"

var shiftAcknowledgementReminderNotificationSubscriptionsCmd = &cobra.Command{
	Use:   "shift-acknowledgement-reminder-notification-subscriptions",
	Short: "View shift acknowledgement reminder notification subscriptions",
	Long: `View shift acknowledgement reminder notification subscriptions on the XBE platform.

These subscriptions determine which users receive reminders to acknowledge shifts.

Commands:
  list    List shift acknowledgement reminder notification subscriptions
  show    Show shift acknowledgement reminder notification subscription details`,
	Example: `  # List subscriptions
  xbe view shift-acknowledgement-reminder-notification-subscriptions list

  # Show a subscription
  xbe view shift-acknowledgement-reminder-notification-subscriptions show 123

  # Output as JSON
  xbe view shift-acknowledgement-reminder-notification-subscriptions list --json`,
}

func init() {
	viewCmd.AddCommand(shiftAcknowledgementReminderNotificationSubscriptionsCmd)
}
