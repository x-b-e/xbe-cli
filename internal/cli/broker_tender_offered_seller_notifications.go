package cli

import "github.com/spf13/cobra"

var brokerTenderOfferedSellerNotificationsCmd = &cobra.Command{
	Use:   "broker-tender-offered-seller-notifications",
	Short: "View broker tender offered seller notifications",
	Long: `View broker tender offered seller notifications on the XBE platform.

These notifications are sent to sellers when a broker tender is offered.

Commands:
  list    List broker tender offered seller notifications
  show    Show broker tender offered seller notification details`,
	Example: `  # List notifications
  xbe view broker-tender-offered-seller-notifications list

  # Show a notification
  xbe view broker-tender-offered-seller-notifications show 123

  # Output as JSON
  xbe view broker-tender-offered-seller-notifications list --json`,
}

func init() {
	viewCmd.AddCommand(brokerTenderOfferedSellerNotificationsCmd)
}
