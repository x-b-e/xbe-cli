package cli

import "github.com/spf13/cobra"

func newServiceTypesShowCmd() *cobra.Command {
	return newGenericShowCmd("service-types")
}

func init() {
	serviceTypesCmd.AddCommand(newServiceTypesShowCmd())
}
