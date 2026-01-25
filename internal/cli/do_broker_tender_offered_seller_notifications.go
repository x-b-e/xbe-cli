package cli

import "github.com/spf13/cobra"

var doBrokerTenderOfferedSellerNotificationsCmd = &cobra.Command{
	Use:   "broker-tender-offered-seller-notifications",
	Short: "Manage broker tender offered seller notifications",
	Long: `Commands for updating broker tender offered seller notifications.

Notifications are created by the platform and can be marked as read by the
recipient.`,
	Example: `  # Mark a notification as read
  xbe do broker-tender-offered-seller-notifications update 123 --read`,
}

func init() {
	doCmd.AddCommand(doBrokerTenderOfferedSellerNotificationsCmd)
}
