package cli

import "github.com/spf13/cobra"

func newLaborersShowCmd() *cobra.Command {
	return newGenericShowCmd("laborers")
}

func init() {
	laborersCmd.AddCommand(newLaborersShowCmd())
}
