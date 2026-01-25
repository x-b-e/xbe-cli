package cli

import "github.com/spf13/cobra"

func newDevelopersShowCmd() *cobra.Command {
	return newGenericShowCmd("developers")
}

func init() {
	developersCmd.AddCommand(newDevelopersShowCmd())
}
