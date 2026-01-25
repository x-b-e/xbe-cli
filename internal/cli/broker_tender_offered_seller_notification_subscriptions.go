package cli

import "github.com/spf13/cobra"

var brokerTenderOfferedSellerNotificationSubscriptionsCmd = &cobra.Command{
	Use:   "broker-tender-offered-seller-notification-subscriptions",
	Short: "View broker tender offered seller notification subscriptions",
	Long: `View broker tender offered seller notification subscriptions on the XBE platform.

These subscriptions determine which users receive notifications when
broker tenders are offered to a trucker.

Commands:
  list    List broker tender offered seller notification subscriptions
  show    Show broker tender offered seller notification subscription details`,
	Example: `  # List subscriptions
  xbe view broker-tender-offered-seller-notification-subscriptions list

  # Show a subscription
  xbe view broker-tender-offered-seller-notification-subscriptions show 123

  # Output as JSON
  xbe view broker-tender-offered-seller-notification-subscriptions list --json`,
}

func init() {
	viewCmd.AddCommand(brokerTenderOfferedSellerNotificationSubscriptionsCmd)
}
