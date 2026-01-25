package cli

import "github.com/spf13/cobra"

func newBrokersShowCmd() *cobra.Command {
	return newGenericShowCmd("brokers")
}

func init() {
	brokersCmd.AddCommand(newBrokersShowCmd())
}
