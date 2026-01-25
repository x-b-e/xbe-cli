package cli

import "github.com/spf13/cobra"

func newRatesShowCmd() *cobra.Command {
	return newGenericShowCmd("rates")
}

func init() {
	ratesCmd.AddCommand(newRatesShowCmd())
}
