package cli

import "github.com/spf13/cobra"

var doCustomerTenderOfferedBuyerNotificationsCmd = &cobra.Command{
	Use:     "customer-tender-offered-buyer-notifications",
	Aliases: []string{"customer-tender-offered-buyer-notification"},
	Short:   "Manage customer tender offered buyer notifications",
	Long:    "Commands for updating customer tender offered buyer notifications.",
}

func init() {
	doCmd.AddCommand(doCustomerTenderOfferedBuyerNotificationsCmd)
}
