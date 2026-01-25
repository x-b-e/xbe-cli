package cli

import "github.com/spf13/cobra"

func newProjectEstimateFileImportsShowCmd() *cobra.Command {
	return newGenericShowCmd("project-estimate-file-imports")
}

func init() {
	projectEstimateFileImportsCmd.AddCommand(newProjectEstimateFileImportsShowCmd())
}
