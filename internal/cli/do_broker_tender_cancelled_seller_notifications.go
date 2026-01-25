package cli

import "github.com/spf13/cobra"

var doBrokerTenderCancelledSellerNotificationsCmd = &cobra.Command{
	Use:     "broker-tender-cancelled-seller-notifications",
	Aliases: []string{"broker-tender-cancelled-seller-notification"},
	Short:   "Manage broker tender cancelled seller notifications",
	Long:    "Commands for updating broker tender cancelled seller notifications.",
}

func init() {
	doCmd.AddCommand(doBrokerTenderCancelledSellerNotificationsCmd)
}
