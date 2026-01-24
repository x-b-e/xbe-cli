package cli

import "github.com/spf13/cobra"

var doTruckerBrokeragesCmd = &cobra.Command{
	Use:     "trucker-brokerages",
	Aliases: []string{"trucker-brokerage"},
	Short:   "Manage trucker brokerages",
	Long:    "Commands for creating, updating, and deleting trucker brokerages.",
}

func init() {
	doCmd.AddCommand(doTruckerBrokeragesCmd)
}
