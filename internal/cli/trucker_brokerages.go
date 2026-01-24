package cli

import "github.com/spf13/cobra"

var truckerBrokeragesCmd = &cobra.Command{
	Use:     "trucker-brokerages",
	Aliases: []string{"trucker-brokerage"},
	Short:   "View trucker brokerages",
	Long:    "Commands for viewing trucker brokerages.",
}

func init() {
	viewCmd.AddCommand(truckerBrokeragesCmd)
}
