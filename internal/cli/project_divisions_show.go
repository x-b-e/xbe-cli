package cli

import "github.com/spf13/cobra"

func newProjectDivisionsShowCmd() *cobra.Command {
	return newGenericShowCmd("project-divisions")
}

func init() {
	projectDivisionsCmd.AddCommand(newProjectDivisionsShowCmd())
}
