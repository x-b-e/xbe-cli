package cli

import "github.com/spf13/cobra"

var brokerTenderCancelledSellerNotificationsCmd = &cobra.Command{
	Use:     "broker-tender-cancelled-seller-notifications",
	Aliases: []string{"broker-tender-cancelled-seller-notification"},
	Short:   "Browse broker tender cancelled seller notifications",
	Long: `Browse broker tender cancelled seller notifications.

Broker tender cancelled seller notifications alert sellers when a broker tender
has been cancelled. Use list to filter and show to view full details.`,
	Example: `  # List recent notifications
  xbe view broker-tender-cancelled-seller-notifications list

  # Filter by read status
  xbe view broker-tender-cancelled-seller-notifications list --read true

  # Show a notification
  xbe view broker-tender-cancelled-seller-notifications show 123`,
}

func init() {
	viewCmd.AddCommand(brokerTenderCancelledSellerNotificationsCmd)
}
