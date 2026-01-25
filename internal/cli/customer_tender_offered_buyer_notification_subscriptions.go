package cli

import "github.com/spf13/cobra"

var customerTenderOfferedBuyerNotificationSubscriptionsCmd = &cobra.Command{
	Use:   "customer-tender-offered-buyer-notification-subscriptions",
	Short: "View customer tender offered buyer notification subscriptions",
	Long: `View customer tender offered buyer notification subscriptions on the XBE platform.

These subscriptions determine which users receive notifications when
customer tenders are offered to a broker.

Commands:
  list    List customer tender offered buyer notification subscriptions
  show    Show customer tender offered buyer notification subscription details`,
	Example: `  # List subscriptions
  xbe view customer-tender-offered-buyer-notification-subscriptions list

  # Show a subscription
  xbe view customer-tender-offered-buyer-notification-subscriptions show 123

  # Output as JSON
  xbe view customer-tender-offered-buyer-notification-subscriptions list --json`,
}

func init() {
	viewCmd.AddCommand(customerTenderOfferedBuyerNotificationSubscriptionsCmd)
}
