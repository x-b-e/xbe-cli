package cli

import "github.com/spf13/cobra"

var doBrokerTenderReturnedBuyerNotificationsCmd = &cobra.Command{
	Use:     "broker-tender-returned-buyer-notifications",
	Aliases: []string{"broker-tender-returned-buyer-notification"},
	Short:   "Manage broker tender returned buyer notifications",
	Long:    "Commands for updating broker tender returned buyer notifications.",
}

func init() {
	doCmd.AddCommand(doBrokerTenderReturnedBuyerNotificationsCmd)
}
