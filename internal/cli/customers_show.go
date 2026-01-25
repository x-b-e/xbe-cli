package cli

import "github.com/spf13/cobra"

func newCustomersShowCmd() *cobra.Command {
	return newGenericShowCmd("customers")
}

func init() {
	customersCmd.AddCommand(newCustomersShowCmd())
}
