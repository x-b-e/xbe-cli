package cli

import "github.com/spf13/cobra"

var brokerTenderReturnedBuyerNotificationsCmd = &cobra.Command{
	Use:     "broker-tender-returned-buyer-notifications",
	Aliases: []string{"broker-tender-returned-buyer-notification"},
	Short:   "Browse broker tender returned buyer notifications",
	Long: `Browse broker tender returned buyer notifications.

Broker tender returned buyer notifications alert buyers when a broker tender
has been returned. Use list to filter and show to view full details.`,
	Example: `  # List recent notifications
  xbe view broker-tender-returned-buyer-notifications list

  # Filter by read status
  xbe view broker-tender-returned-buyer-notifications list --read true

  # Show a notification
  xbe view broker-tender-returned-buyer-notifications show 123`,
}

func init() {
	viewCmd.AddCommand(brokerTenderReturnedBuyerNotificationsCmd)
}
