package cli

import "github.com/spf13/cobra"

func newBusinessUnitsShowCmd() *cobra.Command {
	return newGenericShowCmd("business-units")
}

func init() {
	businessUnitsCmd.AddCommand(newBusinessUnitsShowCmd())
}
