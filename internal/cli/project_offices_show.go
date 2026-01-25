package cli

import "github.com/spf13/cobra"

func newProjectOfficesShowCmd() *cobra.Command {
	return newGenericShowCmd("project-offices")
}

func init() {
	projectOfficesCmd.AddCommand(newProjectOfficesShowCmd())
}
