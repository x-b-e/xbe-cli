package cli

import "github.com/spf13/cobra"

var customerTenderOfferedBuyerNotificationsCmd = &cobra.Command{
	Use:     "customer-tender-offered-buyer-notifications",
	Aliases: []string{"customer-tender-offered-buyer-notification"},
	Short:   "Browse customer tender offered buyer notifications",
	Long: `Browse customer tender offered buyer notifications.

Customer tender offered buyer notifications alert buyers when a customer tender
has been offered. Use list to filter and show to view full details.`,
	Example: `  # List recent notifications
  xbe view customer-tender-offered-buyer-notifications list

  # Filter by read status
  xbe view customer-tender-offered-buyer-notifications list --read true

  # Show a notification
  xbe view customer-tender-offered-buyer-notifications show 123`,
}

func init() {
	viewCmd.AddCommand(customerTenderOfferedBuyerNotificationsCmd)
}
