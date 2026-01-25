package cli

import "github.com/spf13/cobra"

func newProjectPhasesShowCmd() *cobra.Command {
	return newGenericShowCmd("project-phases")
}

func init() {
	projectPhasesCmd.AddCommand(newProjectPhasesShowCmd())
}
