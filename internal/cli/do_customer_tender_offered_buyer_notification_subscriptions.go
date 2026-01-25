package cli

import "github.com/spf13/cobra"

var doCustomerTenderOfferedBuyerNotificationSubscriptionsCmd = &cobra.Command{
	Use:   "customer-tender-offered-buyer-notification-subscriptions",
	Short: "Manage customer tender offered buyer notification subscriptions",
	Long:  "Commands for creating, updating, and deleting customer tender offered buyer notification subscriptions.",
}

func init() {
	doCmd.AddCommand(doCustomerTenderOfferedBuyerNotificationSubscriptionsCmd)
}
