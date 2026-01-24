package cli

import "github.com/spf13/cobra"

var doBrokerTenderOfferedSellerNotificationSubscriptionsCmd = &cobra.Command{
	Use:   "broker-tender-offered-seller-notification-subscriptions",
	Short: "Manage broker tender offered seller notification subscriptions",
	Long:  "Commands for creating, updating, and deleting broker tender offered seller notification subscriptions.",
}

func init() {
	doCmd.AddCommand(doBrokerTenderOfferedSellerNotificationSubscriptionsCmd)
}
