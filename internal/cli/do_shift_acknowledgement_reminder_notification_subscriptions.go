package cli

import "github.com/spf13/cobra"

var doShiftAcknowledgementReminderNotificationSubscriptionsCmd = &cobra.Command{
	Use:   "shift-acknowledgement-reminder-notification-subscriptions",
	Short: "Manage shift acknowledgement reminder notification subscriptions",
	Long:  "Commands for creating, updating, and deleting shift acknowledgement reminder notification subscriptions.",
}

func init() {
	doCmd.AddCommand(doShiftAcknowledgementReminderNotificationSubscriptionsCmd)
}
