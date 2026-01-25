package cli

import "github.com/spf13/cobra"

func newProjectCategoriesShowCmd() *cobra.Command {
	return newGenericShowCmd("project-categories")
}

func init() {
	projectCategoriesCmd.AddCommand(newProjectCategoriesShowCmd())
}
