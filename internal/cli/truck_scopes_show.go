package cli

import "github.com/spf13/cobra"

func newTruckScopesShowCmd() *cobra.Command {
	return newGenericShowCmd("truck-scopes")
}

func init() {
	truckScopesCmd.AddCommand(newTruckScopesShowCmd())
}
