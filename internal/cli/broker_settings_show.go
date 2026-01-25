package cli

import "github.com/spf13/cobra"

func newBrokerSettingsShowCmd() *cobra.Command {
	return newGenericShowCmd("broker-settings")
}

func init() {
	brokerSettingsCmd.AddCommand(newBrokerSettingsShowCmd())
}
